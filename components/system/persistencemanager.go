package system

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/stexxo/dynocue/core/logging"
	"github.com/stexxo/dynocue/core/messaging"
	"golang.org/x/sync/errgroup"
)

type PersistenceManager struct {
	name        string
	kvStore     jetstream.KeyValue
	objectStore jetstream.ObjectStore
	logger      logging.Logger
}

func RegisterWithPersistence(m *messaging.Messenger, logger logging.Logger, subsystemName string, saveSubject string, loadSubject string) (*PersistenceManager, error) {
	resp, err := messaging.RequestRetry[PersistenceRegistrationResponse](m, PersistenceRegistrationRequestSubject, &PersistenceRegistrationRequest{SubsystemName: subsystemName, SaveSubject: saveSubject, LoadSubject: loadSubject}, 10, 500*time.Millisecond)
	if err != nil {
		return nil, err
	}

	kvStore, err := m.JetStream().KeyValue(context.Background(), resp.Response.KeyValueStoreName)
	if err != nil {
		return nil, err
	}

	objectStore, err := m.JetStream().ObjectStore(context.Background(), resp.Response.ObjectStoreName)
	if err != nil {
		return nil, err
	}

	return &PersistenceManager{
		name:        subsystemName,
		kvStore:     kvStore,
		objectStore: objectStore,
		logger:      logger,
	}, nil
}

func (pm *PersistenceManager) KeyValueStore() jetstream.KeyValue {
	return pm.kvStore
}

func (pm *PersistenceManager) ObjectStore() jetstream.ObjectStore {
	return pm.objectStore
}

func (pm *PersistenceManager) WriteToObjectStore(key string, data interface{}) error {
	pm.logger.Info("writing data to store", "key", key, "subsystem", pm.name)

	read, write := io.Pipe()

	execgroup := errgroup.Group{}

	execgroup.Go(func() error {
		defer func() {
			if err := write.Close(); err != nil {
				pm.logger.Error("failed to close pipe writer", "err", err)
			}
		}()

		w := gzip.NewWriter(write)
		defer func() {
			if err := w.Close(); err != nil {
				pm.logger.Error("failed to close gzip writer", "err", err)
			}
		}()

		err := json.NewEncoder(w).Encode(data)
		if err != nil {
			pm.logger.Error("failed to encode model to gzip writer", "err", err)
			return err
		}

		return nil
	})

	execgroup.Go(func() error {
		_, err := pm.objectStore.Put(context.Background(), jetstream.ObjectMeta{Name: fmt.Sprintf("%s/%s", pm.name, key)}, read)
		if err != nil {
			pm.logger.Error("failed to write object to store", "err", err)
		}
		return nil
	})

	err := execgroup.Wait()
	if err != nil {
		return err
	}

	return nil
}

func (pm *PersistenceManager) ReadFromObjectStore(key string, out interface{}) error {
	pm.logger.Info("reading data from store", "key", key, "subsystem", pm.name)

	res, err := pm.objectStore.Get(context.Background(), fmt.Sprintf("%s/%s", pm.name, key))
	if errors.Is(err, jetstream.ErrObjectNotFound) { // nothing to load
		return nil
	}

	if err != nil {
		return err
	}
	defer func() {
		if err := res.Close(); err != nil {
			pm.logger.Error("failed to close object to store", "err", err)
		}
	}()

	readGzip, err := gzip.NewReader(res)
	if err != nil {
		return err
	}
	defer func() {
		if err := readGzip.Close(); err != nil {
			pm.logger.Error("failed to close gzip reader", "err", err)
		}
	}()

	err = json.NewDecoder(readGzip).Decode(out)
	if err != nil {
		return err
	}

	return nil
}
