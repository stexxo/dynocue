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
	Open  string
	Close string
	Save  string
}

const PersistenceKeyValueBucketName = "working-kv"
const PersistenceObjectBucketName = "working-object"

type Persistence struct {
	*core.SubsystemCore

	// Persistence Management
	registeredSubsystems map[string]registeredSubsystem
	kvStore              jetstream.KeyValue
	objectStore          jetstream.ObjectStore

	showOpen bool
	savePath string
}

func NewPersistence(logger logging.Logger) *Persistence {
	p := &Persistence{
		registeredSubsystems: make(map[string]registeredSubsystem),
	}
	p.SubsystemCore = core.NewSubsystemCore("persistence", logger, p.onStart)
	return p
}

func (p *Persistence) onStart() error {
	err := errors.Join(
		messaging.Reply[PersistenceRegistrationRequest, PersistenceRegistrationResponse](p.Messenger(), false, PersistenceRegistrationRequestSubject, p.RegisterRequest),
		messaging.Reply[PersistenceSaveRequest, PersistenceSaveResponse](p.Messenger(), false, PersistenceSaveRequestSubject, p.SaveRequest),
	)

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

	return nil
}

const PersistenceRegistrationRequestSubject = "request.system.persistence.register"

type PersistenceRegistrationRequest struct {
	SubsystemName string `json:"subsystemName" msgpack:"subsystemName"`
	SaveSubject   string `json:"saveSubject" msgpack:"saveSubject"`
	CloseSubject  string `json:"closeSubject" msgpack:"closeSubject"`
	OpenSubject   string `json:"openSubject" msgpack:"openSubject"`
}

type PersistenceRegistrationResponse struct {
	ObjectStoreName   string `json:"objectStoreName" msgpack:"objectStoreName"`
	KeyValueStoreName string `json:"keyValueStoreName" msgpack:"keyValueStoreName"`
}

func (p *Persistence) RegisterRequest(sub string, in *PersistenceRegistrationRequest) (*PersistenceRegistrationResponse, error) {
	p.registeredSubsystems[sub] = registeredSubsystem{Name: in.SubsystemName, Open: in.OpenSubject, Close: in.CloseSubject, Save: in.SaveSubject}
	p.Logger().Debug("registered subsystem for persistence", "subsystem", in.SubsystemName)
	return &PersistenceRegistrationResponse{
		ObjectStoreName:   PersistenceObjectBucketName,
		KeyValueStoreName: PersistenceKeyValueBucketName,
	}, nil
}

type PersistenceOpenRequest struct{}
type PersistenceOpenResponse struct{}

func (p *Persistence) OpenRequest(sub string, in *PersistenceOpenRequest) (*PersistenceOpenResponse, error) {
	if p.showOpen {
		return nil, &messaging.FriendlyError{FriendlyErr: "Cannot Open Show when another is already open. Please close current show before opening"}
	}

	return &PersistenceOpenResponse{}, nil
}

type PersistenceNewRequest struct{}
type PersistenceNewResponse struct{}

func (p *Persistence) NewRequest(sub string, in *PersistenceNewRequest) (*PersistenceNewResponse, error) {
	return nil, nil
}

type PersistenceSaveRequest struct {
	Location string `json:"location" msgpack:"location"`
}
type PersistenceSaveResponse struct{}

const PersistenceSaveRequestSubject = "request.system.persistence.save"

// SaveRequest handles a persistence save request by orchestrating the saving of all registered subsystems
// and then atomically updating the project's zip file.
func (p *Persistence) SaveRequest(sub string, in *PersistenceSaveRequest) (*PersistenceSaveResponse, error) {
	if in.Location == "" && p.savePath == "" {
		return nil, &messaging.FriendlyError{FriendlyErr: "No location provided to save to"}
	}

	// Trigger a save across all registered subsystems.
	p.Logger().Debug("asking subsystems to save to stores")
	execgroup := errgroup.Group{}
	for _, subsystem := range p.registeredSubsystems {
		subsystem := subsystem
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
	newFile, err := os.Create(tempPath)
	if err != nil {
		p.Logger().Error("failed to create temporary file for saving", "path", tempPath, "error", err)
		return nil, &messaging.FriendlyError{FriendlyErr: "failed to create file for saving", Err: err}
	}
	// Ensure the temporary file is removed if the function exits early.
	defer os.Remove(tempPath)

	writer := zip.NewWriter(newFile)

	// Open the existing zip file for reconciliation if it exists.
	oldZip, _ := zip.OpenReader(p.savePath)

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
		oldZip.Close()
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
		stream.Close()
		if err != nil {
			p.Logger().Error("failed to copy stream to zip", "name", name, "error", err)
			return nil, err
		}
	}
	p.Logger().Debug("all new or dirty items added to zip")

	// Finalize the archive and perform the atomic file swap.
	if err := writer.Close(); err != nil {
		p.Logger().Error("failed to close zip writer", "error", err)
		return nil, err
	}
	if err := newFile.Close(); err != nil {
		p.Logger().Error("failed to close new file", "path", tempPath, "error", err)
		return nil, err
	}

	// Renaming the temporary file to the final destination ensures a consistent state.
	err = os.Rename(tempPath, p.savePath)
	if err != nil {
		p.Logger().Error("failed to rename temp file to save path", "tempPath", tempPath, "savePath", p.savePath, "error", err)
		return nil, err
	}
	p.Logger().Debug("save completed successfully", "path", p.savePath)
	return &PersistenceSaveResponse{}, nil
}

type PersistenceCloseRequest struct{}
type PersistenceCloseResponse struct{}

func (p *Persistence) CloseRequest(sub string, in *PersistenceCloseRequest) (*PersistenceCloseResponse, error) {
	return &PersistenceCloseResponse{}, nil
}
