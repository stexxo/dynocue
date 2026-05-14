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

func TestActionTemplateEventsMethods(t *testing.T) {
	api := &CueingApi{}

	t.Run("ActionTemplateCreated returns correct subject and payload", func(t *testing.T) {
		ev := util.Event{
			Operation: model.OperationCreated,
			EventData: map[string]string{
				model.MetadataActionTemplateId: "tmpl-1",
			},
		}
		sub, payload := api.ActionTemplateChanged(ev)
		assert.Equal(t, ActionTemplateCreatedEventSubject, sub)
		assert.Equal(t, "tmpl-1", payload.TemplateId)
	})

	t.Run("ActionTemplateDeleted returns correct subject and payload", func(t *testing.T) {
		ev := util.Event{
			Operation: model.OperationDeleted,
			EventData: map[string]string{
				model.MetadataActionTemplateId: "tmpl-1",
			},
		}
		sub, payload := api.ActionTemplateChanged(ev)
		assert.Equal(t, DeleteActionTemplateEventSubject, sub)
		assert.Equal(t, "tmpl-1", payload.TemplateId)
	})
}

func TestRegisterActionTemplateEvents(t *testing.T) {
	s, nc := testServer()
	defer nc.Close()
	defer s.Shutdown()

	m, err := model.NewCueingModel()
	require.NoError(t, err)

	messenger := messaging.NewMessenger(&messaging.MessengerCfg{Conn: nc})
	_, err = NewCueingApi(m, messenger, nil)
	require.NoError(t, err)

	t.Run("ActionTemplateCreated event is published", func(t *testing.T) {
		wg := sync.WaitGroup{}
		wg.Add(1)
		var receivedId string

		sub, err := nc.Subscribe(ActionTemplateCreatedEventSubject, func(m *nats.Msg) {
			var event ActionTemplateChangeEvent
			_ = msgpack.Unmarshal(m.Data, &event)
			receivedId = event.TemplateId
			wg.Done()
		})
		require.NoError(t, err)
		defer sub.Unsubscribe()

		err = m.RegisterActionTemplate(&types.ActionTemplate{TemplateId: "tmpl-1", TemplateName: "Template 1"})
		require.NoError(t, err)

		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			assert.Equal(t, "tmpl-1", receivedId)
		case <-time.After(time.Second):
			t.Fatal("timed out waiting for ActionTemplateCreated event")
		}
	})

	t.Run("ActionTemplateDeleted event is published", func(t *testing.T) {
		_ = m.RegisterActionTemplate(&types.ActionTemplate{TemplateId: "tmpl-2", TemplateName: "Template 2"})

		wg := sync.WaitGroup{}
		wg.Add(1)
		var receivedId string

		sub, err := nc.Subscribe(DeleteActionTemplateEventSubject, func(m *nats.Msg) {
			var event ActionTemplateChangeEvent
			_ = msgpack.Unmarshal(m.Data, &event)
			receivedId = event.TemplateId
			wg.Done()
		})
		require.NoError(t, err)
		defer sub.Unsubscribe()

		err = m.DeleteActionTemplateById("tmpl-2")
		require.NoError(t, err)

		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			assert.Equal(t, "tmpl-2", receivedId)
		case <-time.After(time.Second):
			t.Fatal("timed out waiting for ActionTemplateDeleted event")
		}
	})
}
