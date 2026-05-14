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
		ev := util.Event{
			Operation: model.OperationCreated,
			EventData: map[string]string{
				model.MetadataCueListId: "cl-1",
				model.MetadataCueId:     "cue-1",
			},
		}
		sub, payload := api.CueChanged(ev)
		assert.Equal(t, CueCreatedEventSubject, sub)
		assert.Equal(t, "cl-1", payload.CueListId)
		assert.Equal(t, "cue-1", payload.CueId)
	})

	t.Run("CueUpdated returns correct subject and payload", func(t *testing.T) {
		ev := util.Event{
			Operation: model.OperationUpdated,
			EventData: map[string]string{
				model.MetadataCueListId: "cl-1",
				model.MetadataCueId:     "cue-1",
			},
		}
		sub, payload := api.CueChanged(ev)
		assert.Equal(t, CueAttributesUpdatedEventSubject, sub)
		assert.Equal(t, "cl-1", payload.CueListId)
		assert.Equal(t, "cue-1", payload.CueId)
	})

	t.Run("CueDeleted returns correct subject and payload", func(t *testing.T) {
		ev := util.Event{
			Operation: model.OperationDeleted,
			EventData: map[string]string{
				model.MetadataCueListId: "cl-1",
				model.MetadataCueId:     "cue-1",
			},
		}
		sub, payload := api.CueChanged(ev)
		assert.Equal(t, DeleteCueEventSubject, sub)
		assert.Equal(t, "cl-1", payload.CueListId)
		assert.Equal(t, "cue-1", payload.CueId)
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

	// Seed data
	clId, _, _ := m.CreateCueList(1, types.CueListTypeSequential)

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

		cueId, _, err := m.CreateCue(clId, 0)
		require.NoError(t, err)

		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			assert.Equal(t, cueId, receivedId)
		case <-time.After(time.Second):
			t.Fatal("timed out waiting for CueCreated event")
		}
	})

	t.Run("CueUpdated event is published", func(t *testing.T) {
		cueId, _, _ := m.CreateCue(clId, 0)

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

		err = m.UpdateCueAttribute(cueId, "label", "new label")
		require.NoError(t, err)

		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			assert.Equal(t, cueId, receivedId)
		case <-time.After(time.Second):
			t.Fatal("timed out waiting for CueUpdated event")
		}
	})

	t.Run("CueDeleted event is published", func(t *testing.T) {
		cueId, _, _ := m.CreateCue(clId, 0)

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

		err = m.DeleteCueById(cueId)
		require.NoError(t, err)

		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			assert.Equal(t, cueId, receivedId)
		case <-time.After(time.Second):
			t.Fatal("timed out waiting for CueDeleted event")
		}
	})
}
