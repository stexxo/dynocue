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
		ev := util.Event{Identifier: "test-id"}
		sub, payload := api.CueListCreated(ev)
		assert.Equal(t, CueListCreatedEventSubject, sub)
		assert.Equal(t, "test-id", payload.CueListId)
	})

	t.Run("CueListUpdated returns correct subject and payload", func(t *testing.T) {
		ev := util.Event{Identifier: "test-id"}
		sub, payload := api.CueListUpdated(ev)
		assert.Equal(t, CueListAttributesUpdatedEventSubject, sub)
		assert.Equal(t, "test-id", payload.CueListId)
	})

	t.Run("DeleteCueListEvent returns correct subject and payload", func(t *testing.T) {
		ev := util.Event{Identifier: "test-id"}
		sub, payload := api.DeleteCueListEvent(ev)
		assert.Equal(t, DeleteCueListEventSubject, sub)
		assert.Equal(t, "test-id", payload.CueListId)
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

	t.Run("Created event is published", func(t *testing.T) {
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

		id, _, err := m.CreateCueList(1, types.CueListTypeSequential)
		require.NoError(t, err)

		// Wait for event with timeout
		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			assert.Equal(t, id, receivedId)
		case <-time.After(time.Second):
			t.Fatal("timed out waiting for created event")
		}
	})

	t.Run("Updated event is published", func(t *testing.T) {
		id, _, _ := m.CreateCueList(2, types.CueListTypeSequential)

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

		err = m.UpdateCueListAttribute(id, "label", "new label")
		require.NoError(t, err)

		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			assert.Equal(t, id, receivedId)
		case <-time.After(time.Second):
			t.Fatal("timed out waiting for updated event")
		}
	})

	t.Run("Deleted event is published", func(t *testing.T) {
		id, _, _ := m.CreateCueList(3, types.CueListTypeSequential)

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

		err = m.DeleteCueListById(id)
		require.NoError(t, err)

		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			assert.Equal(t, id, receivedId)
		case <-time.After(time.Second):
			t.Fatal("timed out waiting for deleted event")
		}
	})
}
