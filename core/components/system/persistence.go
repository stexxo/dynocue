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
	"github.com/stexxo/dynocue/core/components"
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
	*components.BaseComponent

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
	p.BaseComponent = components.NewBaseComponent("persistence", logger, p.onStart)
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

func (p *Persistence) SaveRequest(sub string, in *PersistenceSaveRequest) (*PersistenceSaveResponse, error) {
	if in.Location == "" && p.savePath == "" {
		return nil, &messaging.FriendlyError{FriendlyErr: "No location provided to save to"}
	}

	// Ask all Subsystems to Save
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
		return nil, err
	}

	p.savePath = cmp.Or(in.Location, p.savePath)

	// 2. Prepare the Atomic Swap (Temp file)
	tempPath := p.savePath + ".tmp"
	newFile, err := os.Create(tempPath)
	if err != nil {
		return nil, &messaging.FriendlyError{FriendlyErr: "failed to create file for saving", Err: err}
	}
	// We'll close this manually, but defer a cleanup in case of panic/early return
	defer os.Remove(tempPath)

	writer := zip.NewWriter(newFile)

	// 3. Open the OLD zip for incremental comparison
	// If it doesn't exist (New Project), we proceed with oldZip as nil
	oldZip, _ := zip.OpenReader(p.savePath)

	// 4. Access Live NATS Inventory (The Source of Truth)
	kv := p.kvStore
	obs := p.objectStore

	// Get maps of what is currently in NATS
	liveKeys := make(map[string]jetstream.KeyValueEntry)
	keys, _ := kv.ListKeys(context.Background())
	if keys != nil {
		for entry := range keys.Keys() {
			liveKeys[entry], err = kv.Get(context.Background(), entry)
			if err != nil {
				return nil, &messaging.FriendlyError{FriendlyErr: "Failed to retrieve keys from Key Value Store.", Err: err}
			}
		}
	}

	liveBlobs := make(map[string]*jetstream.ObjectInfo)
	blobs, _ := obs.List(context.Background())
	for _, info := range blobs {
		liveBlobs[info.Name] = info
	}

	// 5. RECONCILIATION: Copy existing valid data from the old Zip
	if oldZip != nil {
		for _, f := range oldZip.File {
			if strings.HasPrefix(f.Name, "kv/") {
				key := strings.TrimPrefix(f.Name, "kv/")
				if entry, exists := liveKeys[key]; exists {
					// Check if Revision matches the Zip Comment
					if f.Comment == fmt.Sprintf("%d", entry.Revision()) {
						writer.Copy(f) // Fast byte-copy
						delete(liveKeys, key)
					}
				}
			} else if strings.HasPrefix(f.Name, "blobs/") {
				name := strings.TrimPrefix(f.Name, "blobs/")
				if info, exists := liveBlobs[name]; exists {
					// Check if SHA-256 Digest matches the Zip Comment
					if f.Comment == info.Digest {
						writer.Copy(f) // Fast byte-copy
						delete(liveBlobs, name)
					}
				}
			}
		}
		oldZip.Close()
	}

	// 6. SHEPHERD: Add remaining (New or Dirty) items to the Zip
	// Shepherd Key-Value Pairs
	for key, entry := range liveKeys {
		header := &zip.FileHeader{
			Name:    "kv/" + key,
			Comment: fmt.Sprintf("%d", entry.Revision()),
		}
		zf, _ := writer.CreateHeader(header)
		zf.Write(entry.Value())
	}

	// Shepherd Large Blobs (100MB chunks)
	for name, info := range liveBlobs {
		header := &zip.FileHeader{
			Name:    "blobs/" + name,
			Method:  zip.Store,   // Speed over compression for binary blobs
			Comment: info.Digest, // Crucial for next sync
		}
		zf, _ := writer.CreateHeader(header)

		// Stream directly from NATS to Zip to keep memory flat
		stream, err := obs.Get(context.Background(), name)
		if err != nil {
			return nil, err
		}
		_, err = io.Copy(zf, stream)
		stream.Close()
		if err != nil {
			return nil, err
		}
	}

	// 7. FINALIZE: Close Zip and Swap
	if err := writer.Close(); err != nil {
		return nil, err
	}
	if err := newFile.Close(); err != nil {
		return nil, err
	}

	// The Atomic Swap: This ensures the project file is never in a "half-written" state
	err = os.Rename(tempPath, p.savePath)
	if err != nil {
		return nil, err
	}
	return &PersistenceSaveResponse{}, nil
}

type PersistenceCloseRequest struct{}
type PersistenceCloseResponse struct{}

func (p *Persistence) CloseRequest(sub string, in *PersistenceCloseRequest) (*PersistenceCloseResponse, error) {
	return &PersistenceCloseResponse{}, nil
}
