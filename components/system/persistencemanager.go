// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package system

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/stexxo/dynocue/core/logging"
	"github.com/stexxo/dynocue/core/messaging"
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

func (pm *PersistenceManager) WriteToObjectStore(key string, reader io.Reader) error {
	key = fmt.Sprintf("%s/%s", pm.name, key)
	pm.logger.Info("writing data to store", "key", key, "subsystem", pm.name)
	_, err := pm.objectStore.Put(context.Background(), jetstream.ObjectMeta{Name: key}, reader)
	if err != nil {
		pm.logger.Error("failed to write object to store", "err", err)
		return err
	}
	return nil
}

func (pm *PersistenceManager) ReadFromObjectStore(key string) (jetstream.ObjectResult, error) {
	key = fmt.Sprintf("%s/%s", pm.name, key)
	pm.logger.Info("reading data from store", "key", key, "subsystem", pm.name)

	res, err := pm.objectStore.Get(context.Background(), key)
	if errors.Is(err, jetstream.ErrObjectNotFound) { // nothing to load
		return nil, err
	}
	if err != nil {
		return nil, err
	}

	return res, nil
}
