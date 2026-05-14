package api

import (
	"sync"
	"testing"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/stexxo/dynocue/components/cues/model"
	"github.com/stexxo/dynocue/components/cues/types"
	"github.com/stexxo/dynocue/core/messaging"
	"github.com/stexxo/dynocue/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vmihailenco/msgpack/v5"
)

func TestCueListEventsMethods(t *testing.T) {
	api := &CueingApi{}

	t.Run("CueListCreated returns correct subject and payload", func(t *testing.T) {
		ev := util.Event{
			Operation: model.OperationCreated,
			EventData: map[string]string{
				model.MetadataCueListId: "cl-1",
			},
		}
		sub, payload := api.CueListChanged(ev)
		assert.Equal(t, CueListCreatedEventSubject, sub)
		assert.Equal(t, "cl-1", payload.CueListId)
	})

	t.Run("CueListUpdated returns correct subject and payload", func(t *testing.T) {
		ev := util.Event{
			Operation: model.OperationUpdated,
			EventData: map[string]string{
				model.MetadataCueListId: "cl-1",
			},
		}
		sub, payload := api.CueListChanged(ev)
		assert.Equal(t, CueListAttributesUpdatedEventSubject, sub)
		assert.Equal(t, "cl-1", payload.CueListId)
	})

	t.Run("CueListDeleted returns correct subject and payload", func(t *testing.T) {
		ev := util.Event{
			Operation: model.OperationDeleted,
			EventData: map[string]string{
				model.MetadataCueListId: "cl-1",
			},
		}
		sub, payload := api.CueListChanged(ev)
		assert.Equal(t, DeleteCueListEventSubject, sub)
		assert.Equal(t, "cl-1", payload.CueListId)
	})
}

func TestRegisterCueListEvents(t *testing.T) {
	s, nc := testServer()
	defer nc.Close()
	defer s.Shutdown()

	m, err := model.NewCueingModel()
	require.NoError(t, err)

	messenger := messaging.NewMessenger(&messaging.MessengerCfg{Conn: nc})
	_, err = NewCueingApi(m, messenger, nil)
	require.NoError(t, err)

	t.Run("CueListCreated event is published", func(t *testing.T) {
		wg := sync.WaitGroup{}
		wg.Add(1)
		var receivedId string

		sub, err := nc.Subscribe(CueListCreatedEventSubject, func(m *nats.Msg) {
			var event CueListChangeEvent
			_ = msgpack.Unmarshal(m.Data, &event)
			receivedId = event.CueListId
			wg.Done()
		})
		require.NoError(t, err)
		defer sub.Unsubscribe()

		clId, _, err := m.CreateCueList(1, types.CueListTypeSequential)
		require.NoError(t, err)

		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			assert.Equal(t, clId, receivedId)
		case <-time.After(time.Second):
			t.Fatal("timed out waiting for CueListCreated event")
		}
	})

	t.Run("CueListUpdated event is published", func(t *testing.T) {
		clId, _, _ := m.CreateCueList(2, types.CueListTypeSequential)

		wg := sync.WaitGroup{}
		wg.Add(1)
		var receivedId string

		sub, err := nc.Subscribe(CueListAttributesUpdatedEventSubject, func(m *nats.Msg) {
			var event CueListChangeEvent
			_ = msgpack.Unmarshal(m.Data, &event)
			receivedId = event.CueListId
			wg.Done()
		})
		require.NoError(t, err)
		defer sub.Unsubscribe()

		err = m.UpdateCueListAttribute(clId, "label", "new label")
		require.NoError(t, err)

		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			assert.Equal(t, clId, receivedId)
		case <-time.After(time.Second):
			t.Fatal("timed out waiting for CueListUpdated event")
		}
	})

	t.Run("CueListDeleted event is published", func(t *testing.T) {
		clId, _, _ := m.CreateCueList(3, types.CueListTypeSequential)

		wg := sync.WaitGroup{}
		wg.Add(1)
		var receivedId string

		sub, err := nc.Subscribe(DeleteCueListEventSubject, func(m *nats.Msg) {
			var event CueListChangeEvent
			_ = msgpack.Unmarshal(m.Data, &event)
			receivedId = event.CueListId
			wg.Done()
		})
		require.NoError(t, err)
		defer sub.Unsubscribe()

		err = m.DeleteCueListById(clId)
		require.NoError(t, err)

		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			assert.Equal(t, clId, receivedId)
		case <-time.After(time.Second):
			t.Fatal("timed out waiting for CueListDeleted event")
		}
	})
}
