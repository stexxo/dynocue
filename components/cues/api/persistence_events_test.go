package api

import (
	"bytes"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/stexxo/dynocue/components/cues/model"
	"github.com/stexxo/dynocue/core/messaging"
	"github.com/stexxo/dynocue/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vmihailenco/msgpack/v5"
)

func TestPersistenceEventsMethods(t *testing.T) {
	api := &CueingApi{}

	t.Run("PersistenceChanged returns correct subject for Loaded", func(t *testing.T) {
		ev := util.Event{
			Operation: model.OperationLoaded,
		}
		sub, payload := api.PersistenceChanged(ev)
		assert.Equal(t, ModelLoadedEventSubject, sub)
		assert.NotNil(t, payload)
	})
}

func TestRegisterPersistenceEvents(t *testing.T) {
	s, nc := testServer()
	defer nc.Close()
	defer s.Shutdown()

	m, err := model.NewCueingModel()
	require.NoError(t, err)

	messenger := messaging.NewMessenger(&messaging.MessengerCfg{Conn: nc})
	_, err = NewCueingApi(m, nil, messenger, nil)
	require.NoError(t, err)

	t.Run("ModelLoaded event is published", func(t *testing.T) {
		wg := sync.WaitGroup{}
		wg.Add(1)

		sub, err := nc.Subscribe(ModelLoadedEventSubject, func(m *nats.Msg) {
			var event PersistenceChangeEvent
			_ = msgpack.Unmarshal(m.Data, &event)
			wg.Done()
		})
		require.NoError(t, err)
		defer sub.Unsubscribe()

		// Mock loader function
		loader := func(name string) (io.Reader, error) {
			return bytes.NewReader([]byte{0x90}), nil // 0x90 is empty array in msgpack
		}

		_ = m.LoadModel(loader)

		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			// Success
		case <-time.After(time.Second):
			t.Fatal("timed out waiting for ModelLoaded event")
		}
	})
}
