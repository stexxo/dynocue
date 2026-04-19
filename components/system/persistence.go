// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package system

import (
	"archive/zip"
	"cmp"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/stexxo/dynocue/core"
	"github.com/stexxo/dynocue/core/logging"
	"github.com/stexxo/dynocue/core/messaging"
	"golang.org/x/sync/errgroup"
)

type registeredSubsystem struct {
	Name  string
	Load  string
	Close string
	Save  string
}

const PersistenceKeyValueBucketName = "working-kv"
const PersistenceObjectBucketName = "working-object"

type Persistence struct {
	*core.SubsystemCore

	// Persistence Management
	registeredSubsystems []registeredSubsystem
	kvStore              jetstream.KeyValue
	objectStore          jetstream.ObjectStore

	showOpen bool
	savePath string
}

func NewPersistence(logger logging.Logger) *Persistence {
	p := &Persistence{}
	p.SubsystemCore = core.NewSubsystemCore("persistence", logger, p.onStart)
	return p
}

func (p *Persistence) onStart() error {
	// Build Object Store and Key Value Store to be used for persistence management
	kv, err := p.Messenger().JetStream().CreateKeyValue(context.Background(), jetstream.KeyValueConfig{
		Bucket:  PersistenceKeyValueBucketName,
		Storage: jetstream.FileStorage,
	})
	if err != nil {
		return err
	}
	p.kvStore = kv

	object, err := p.Messenger().JetStream().CreateObjectStore(context.Background(), jetstream.ObjectStoreConfig{
		Bucket:  PersistenceObjectBucketName,
		Storage: jetstream.FileStorage,
	})
	if err != nil {
		return err
	}
	p.objectStore = object

	err = p.softClear()
	if err != nil {
		return err
	}

	err = errors.Join(
		messaging.Reply[PersistenceRegistrationRequest, PersistenceRegistrationResponse](p.Messenger(), false, PersistenceRegistrationRequestSubject, p.RegisterRequest),
		messaging.Reply[PersistenceSaveRequest, PersistenceSaveResponse](p.Messenger(), false, PersistenceSaveRequestSubject, p.SaveRequest),
		messaging.Reply[PersistenceOpenRequest, PersistenceOpenResponse](p.Messenger(), false, PersistenceOpenShowRequestSubject, p.OpenRequest),
		messaging.Reply[PersistenceNewRequest, PersistenceNewResponse](p.Messenger(), false, PersistenceNewShowRequestSubject, p.NewRequest),
	)

	return nil
}

func (p *Persistence) softClear() error {
	keys, _ := p.kvStore.ListKeys(context.Background())
	if keys != nil {
		for entry := range keys.Keys() {
			err := p.kvStore.Delete(context.Background(), entry)
			if err != nil {
				return err
			}
		}
	}

	objs, _ := p.objectStore.List(context.Background())
	for _, info := range objs {
		err := p.objectStore.Delete(context.Background(), info.Name)
		if err != nil {
			return err
		}
	}

	return nil
}

const PersistenceShowLoadedEventSubject = "event.system.persistence.loaded"
const PersistenceShowSavedEventSubject = "event.system.persistence.saved"

const PersistenceRegistrationRequestSubject = "request.system.persistence.register"

type PersistenceRegistrationRequest struct {
	SubsystemName string `json:"subsystemName" msgpack:"subsystemName"`
	SaveSubject   string `json:"saveSubject" msgpack:"saveSubject"`
	LoadSubject   string `json:"loadSubject" msgpack:"loadSubject"`
}

type PersistenceRegistrationResponse struct {
	ObjectStoreName   string `json:"objectStoreName" msgpack:"objectStoreName"`
	KeyValueStoreName string `json:"keyValueStoreName" msgpack:"keyValueStoreName"`
}

func (p *Persistence) RegisterRequest(sub string, in *PersistenceRegistrationRequest) (*PersistenceRegistrationResponse, error) {
	p.registeredSubsystems = append(p.registeredSubsystems, registeredSubsystem{Name: in.SubsystemName, Load: in.LoadSubject, Save: in.SaveSubject})
	p.Logger().Debug("registered subsystem for persistence", "subsystem", in.SubsystemName)
	return &PersistenceRegistrationResponse{
		ObjectStoreName:   PersistenceObjectBucketName,
		KeyValueStoreName: PersistenceKeyValueBucketName,
	}, nil
}

const PersistenceOpenShowRequestSubject = "event.system.persistence.open"

type PersistenceOpenRequest struct {
	Location string `json:"location" msgpack:"location"`
}
type PersistenceOpenResponse struct{}

func (p *Persistence) OpenRequest(sub string, in *PersistenceOpenRequest) (*PersistenceOpenResponse, error) {
	if _, err := os.Stat(in.Location); os.IsNotExist(err) {
		return nil, &messaging.FriendlyError{FriendlyErr: "Provided show location does not exist."}
	}

	err := p.softClear()
	if err != nil {
		return nil, &messaging.FriendlyError{FriendlyErr: "failed to clear object and key value store"}
	}

	zr, err := zip.OpenReader(in.Location)
	if err != nil {
		return nil, &messaging.FriendlyError{FriendlyErr: "failed to open archive"}
	}
	defer func() {
		err := zr.Close()
		if err != nil {
			p.Logger().Error("failed to close archive reader", "error", err)
		}
	}()

	for _, f := range zr.File {
		// Skip directories or unrelated files
		if f.FileInfo().IsDir() {
			continue
		}

		rc, err := f.Open()
		if err != nil {
			return nil, err
		}

		if strings.HasPrefix(f.Name, "kv/") {
			key := strings.TrimPrefix(f.Name, "kv/")
			data, _ := io.ReadAll(rc)
			_, err = p.kvStore.Put(context.Background(), key, data)
		} else if strings.HasPrefix(f.Name, "blobs/") {
			name := strings.TrimPrefix(f.Name, "blobs/")
			_, err = p.objectStore.Put(context.Background(), jetstream.ObjectMeta{Name: name}, rc)
		}

		closeErr := rc.Close()
		if closeErr != nil {
			p.Logger().Error("failed to close archive reader", "error", err)
		}

		if err != nil {
			return nil, fmt.Errorf("failed to rehydrate %s: %w", f.Name, err)
		}
	}

	p.showOpen = true
	p.savePath = in.Location

	p.Logger().Debug("asking subsystems to reload stores")
	execgroup := errgroup.Group{}
	for _, subsystem := range p.registeredSubsystems {
		execgroup.Go(func() error {
			_, err := messaging.Request[string](p.Messenger(), subsystem.Load, "")
			return err
		})
	}

	err = execgroup.Wait()
	if err != nil {
		return nil, err
	}

	err = messaging.Publish(p.Messenger(), PersistenceShowLoadedEventSubject, "")
	if err != nil {
		return nil, err
	}

	return &PersistenceOpenResponse{}, nil
}

const PersistenceNewShowRequestSubject = "event.system.persistence.new"

type PersistenceNewRequest struct{}
type PersistenceNewResponse struct{}

func (p *Persistence) NewRequest(sub string, in *PersistenceNewRequest) (*PersistenceNewResponse, error) {
	err := p.softClear()
	if err != nil {
		return nil, &messaging.FriendlyError{FriendlyErr: "failed to clear object"}
	}

	p.Logger().Debug("asking subsystems to reload store")
	execgroup := errgroup.Group{}
	for _, subsystem := range p.registeredSubsystems {
		execgroup.Go(func() error {
			_, err := messaging.Request[string](p.Messenger(), subsystem.Load, "")
			return err
		})
	}

	err = execgroup.Wait()
	if err != nil {
		return nil, err
	}

	err = messaging.Publish(p.Messenger(), PersistenceShowLoadedEventSubject, "")
	if err != nil {
		return nil, err
	}

	return &PersistenceNewResponse{}, nil
}

type PersistenceSaveRequest struct {
	Location string `json:"location" msgpack:"location"`
}
type PersistenceSaveResponse struct{}

const PersistenceSaveRequestSubject = "request.system.persistence.save"

const NoSaveLocation string = "No Save Location Provided."

// SaveRequest handles a persistence save request by orchestrating the saving of all registered subsystems
// and then atomically updating the project's zip file.
func (p *Persistence) SaveRequest(sub string, in *PersistenceSaveRequest) (*PersistenceSaveResponse, error) {
	if in.Location == "" && p.savePath == "" {
		return nil, &messaging.FriendlyError{FriendlyErr: NoSaveLocation}
	}

	if !strings.HasSuffix(in.Location, ".dyno") {
		in.Location = in.Location + ".dyno"
	}

	// Trigger a save across all registered subsystems.
	p.Logger().Debug("asking subsystems to save to stores")
	execgroup := errgroup.Group{}
	for _, subsystem := range p.registeredSubsystems {
		execgroup.Go(func() error {
			_, err := messaging.Request[string](p.Messenger(), subsystem.Save, "")
			return err
		})
	}

	p.Logger().Debug("waiting for subsystem replies")
	if err := execgroup.Wait(); err != nil {
		p.Logger().Error("subsystems failed to save", "error", err)
		return nil, err
	}
	p.Logger().Debug("subsystems saved successfully")

	p.savePath = cmp.Or(in.Location, p.savePath)

	// Prepare for an atomic swap by writing to a temporary file.
	tempPath := p.savePath + ".tmp"
	defer func() {
		err := os.Remove(tempPath)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			p.Logger().Error("failed to remove temp file", "error", err)
		}
	}()

	newFile, err := os.Create(tempPath)
	if err != nil {
		p.Logger().Error("failed to create temporary file for saving", "path", tempPath, "error", err)
		return nil, &messaging.FriendlyError{FriendlyErr: "failed to create file for saving", Err: err}
	}
	defer func() {
		if err := newFile.Close(); err != nil {
			p.Logger().Error("failed to close new file", "path", tempPath, "error", err)
		}
	}()

	writer := zip.NewWriter(newFile)
	defer func() {
		if err := writer.Close(); err != nil {
			p.Logger().Error("failed to close zip writer", "error", err)
		}
	}()

	// Load the existing zip file for reconciliation if it exists.
	oldZip, _ := zip.OpenReader(p.savePath)
	defer func() {
		if oldZip != nil {
			if err := oldZip.Close(); err != nil {
				p.Logger().Error("failed to close old zip", "path", p.savePath, "error", err)
			}
		}
	}()

	// Access the current state from the NATS inventory.
	kv := p.kvStore
	obs := p.objectStore

	// Retrieve all current keys from the Key-Value store.
	liveKeys := make(map[string]jetstream.KeyValueEntry)
	keys, _ := kv.ListKeys(context.Background())
	if keys != nil {
		for entry := range keys.Keys() {
			liveKeys[entry], err = kv.Get(context.Background(), entry)
			if err != nil {
				p.Logger().Error("failed to retrieve key from Key Value Store", "key", entry, "error", err)
				return nil, &messaging.FriendlyError{FriendlyErr: "Failed to retrieve keys from Key Value Store.", Err: err}
			}
		}
	}

	// Map all current objects in the object store.
	liveBlobs := make(map[string]*jetstream.ObjectInfo)
	blobs, _ := obs.List(context.Background())
	for _, info := range blobs {
		liveBlobs[info.Name] = info
	}

	// Reconciliation: Reuse valid data from the existing zip file if possible.
	if oldZip != nil {
		for _, f := range oldZip.File {
			if strings.HasPrefix(f.Name, "kv/") {
				key := strings.TrimPrefix(f.Name, "kv/")
				if entry, exists := liveKeys[key]; exists {
					// Verify if the revision matches the zip entry's comment.
					if f.Comment == fmt.Sprintf("%d", entry.Revision()) {
						if err := writer.Copy(f); err != nil {
							p.Logger().Error("failed to copy file from old zip to new zip", "file", f.Name, "error", err)
							return nil, err
						}
						delete(liveKeys, key)
					}
				}
			} else if strings.HasPrefix(f.Name, "blobs/") {
				name := strings.TrimPrefix(f.Name, "blobs/")
				if info, exists := liveBlobs[name]; exists {
					// Verify if the SHA-256 digest matches the zip entry's comment.
					if f.Comment == info.Digest {
						if err := writer.Copy(f); err != nil {
							p.Logger().Error("failed to copy blob from old zip to new zip", "file", f.Name, "error", err)
							return nil, err
						}
						delete(liveBlobs, name)
					}
				}
			}
		}
		p.Logger().Debug("reconciliation complete: reused existing valid data from old zip")
	}

	// Append any remaining new or modified items to the new zip file.
	// Process Key-Value pairs.
	for key, entry := range liveKeys {
		header := &zip.FileHeader{
			Name:    "kv/" + key,
			Comment: fmt.Sprintf("%d", entry.Revision()),
		}
		zf, err := writer.CreateHeader(header)
		if err != nil {
			p.Logger().Error("failed to create zip header for key", "key", key, "error", err)
			return nil, err
		}
		if _, err := zf.Write(entry.Value()); err != nil {
			p.Logger().Error("failed to write key value to zip", "key", key, "error", err)
			return nil, err
		}
	}

	// Process large blobs, streaming them directly from NATS.
	for name, info := range liveBlobs {
		header := &zip.FileHeader{
			Name:    "blobs/" + name,
			Method:  zip.Store,   // Optimization: Skip compression for binary blobs.
			Comment: info.Digest, // Used for reconciliation in subsequent saves.
		}
		zf, err := writer.CreateHeader(header)
		if err != nil {
			p.Logger().Error("failed to create zip header for blob", "name", name, "error", err)
			return nil, err
		}

		// Stream content directly from the object store to minimize memory usage.
		stream, err := obs.Get(context.Background(), name)
		if err != nil {
			p.Logger().Error("failed to get stream from object store", "name", name, "error", err)
			return nil, err
		}
		_, err = io.Copy(zf, stream)

		if err := stream.Close(); err != nil {
			p.Logger().Error("failed to close stream", "name", name, "error", err)
			return nil, err
		}

		if err != nil {
			p.Logger().Error("failed to copy stream to zip", "name", name, "error", err)
			return nil, err
		}
	}
	p.Logger().Debug("all new or dirty items added to zip")

	// Renaming the temporary file to the final destination ensures a consistent state.
	err = os.Rename(tempPath, p.savePath)
	if err != nil {
		p.Logger().Error("failed to rename temp file to save path", "tempPath", tempPath, "savePath", p.savePath, "error", err)
		return nil, err
	}

	err = messaging.Publish[string](p.Messenger(), PersistenceShowSavedEventSubject, p.savePath)
	if err != nil {
		p.Logger().Error("failed to publish saved event", "error", err)
	}

	p.Logger().Debug("save completed successfully", "path", p.savePath)

	return &PersistenceSaveResponse{}, nil
}
