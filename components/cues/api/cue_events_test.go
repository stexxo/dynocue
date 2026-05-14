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

func TestCueEventsMethods(t *testing.T) {
	api := &CueingApi{}

	t.Run("CueCreated returns correct subject and payload", func(t *testing.T) {
		ev := util.Event{Identifier: "test-cue-id"}
		sub, payload := api.CueCreated(ev)
		assert.Equal(t, CueCreatedEventSubject, sub)
		assert.Equal(t, "test-cue-id", payload.CueId)
	})

	t.Run("CueUpdated returns correct subject and payload", func(t *testing.T) {
		ev := util.Event{Identifier: "test-cue-id"}
		sub, payload := api.CueUpdated(ev)
		assert.Equal(t, CueAttributesUpdatedEventSubject, sub)
		assert.Equal(t, "test-cue-id", payload.CueId)
	})

	t.Run("DeleteCueEvent returns correct subject and payload", func(t *testing.T) {
		ev := util.Event{Identifier: "test-cue-id"}
		sub, payload := api.DeleteCueEvent(ev)
		assert.Equal(t, DeleteCueEventSubject, sub)
		assert.Equal(t, "test-cue-id", payload.CueId)
	})
}

func TestRegisterCueEvents(t *testing.T) {
	s, nc := testServer()
	defer nc.Close()
	defer s.Shutdown()

	m, err := model.NewCueingModel()
	require.NoError(t, err)

	messenger := messaging.NewMessenger(&messaging.MessengerCfg{Conn: nc})
	_, err = NewCueingApi(m, messenger, nil)
	require.NoError(t, err)

	// Create a cue list first since cues need one
	clId, _, err := m.CreateCueList(1, types.CueListTypeSequential)
	require.NoError(t, err)

	t.Run("CueCreated event is published", func(t *testing.T) {
		wg := sync.WaitGroup{}
		wg.Add(1)
		var receivedId string

		sub, err := nc.Subscribe(CueCreatedEventSubject, func(m *nats.Msg) {
			var event CueChangeEvent
			_ = msgpack.Unmarshal(m.Data, &event)
			receivedId = event.CueId
			wg.Done()
		})
		require.NoError(t, err)
		defer sub.Unsubscribe()

		id, _, err := m.CreateCue(clId, 1)
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
			t.Fatal("timed out waiting for cue created event")
		}
	})

	t.Run("CueUpdated event is published", func(t *testing.T) {
		id, _, _ := m.CreateCue(clId, 2)

		wg := sync.WaitGroup{}
		wg.Add(1)
		var receivedId string

		sub, err := nc.Subscribe(CueAttributesUpdatedEventSubject, func(m *nats.Msg) {
			var event CueChangeEvent
			_ = msgpack.Unmarshal(m.Data, &event)
			receivedId = event.CueId
			wg.Done()
		})
		require.NoError(t, err)
		defer sub.Unsubscribe()

		err = m.UpdateCueAttribute(id, "label", "new cue label")
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
			t.Fatal("timed out waiting for cue updated event")
		}
	})

	t.Run("CueDeleted event is published", func(t *testing.T) {
		id, _, _ := m.CreateCue(clId, 3)

		wg := sync.WaitGroup{}
		wg.Add(1)
		var receivedId string

		sub, err := nc.Subscribe(DeleteCueEventSubject, func(m *nats.Msg) {
			var event CueChangeEvent
			_ = msgpack.Unmarshal(m.Data, &event)
			receivedId = event.CueId
			wg.Done()
		})
		require.NoError(t, err)
		defer sub.Unsubscribe()

		err = m.DeleteCueById(id)
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
			t.Fatal("timed out waiting for cue deleted event")
		}
	})
}
