// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

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

func TestActionEventsMethods(t *testing.T) {
	api := &CueingApi{}

	t.Run("ActionCreated returns correct subject and payload", func(t *testing.T) {
		ev := util.Event{
			Operation: model.OperationCreated,
			EventData: map[string]string{
				model.MetadataCueListId: "cl-1",
				model.MetadataCueId:     "cue-1",
				model.MetadataActionId:  "act-1",
			},
		}
		sub, payload := api.ActionChanged(ev)
		assert.Equal(t, ActionCreatedEventSubject, sub)
		assert.Equal(t, "cl-1", payload.CueListId)
		assert.Equal(t, "cue-1", payload.CueId)
		assert.Equal(t, "act-1", payload.ActionId)
	})

	t.Run("ActionUpdated returns correct subject and payload", func(t *testing.T) {
		ev := util.Event{
			Operation: model.OperationUpdated,
			EventData: map[string]string{
				model.MetadataCueListId: "cl-1",
				model.MetadataCueId:     "cue-1",
				model.MetadataActionId:  "act-1",
			},
		}
		sub, payload := api.ActionChanged(ev)
		assert.Equal(t, ActionAttributesUpdatedEventSubject, sub)
		assert.Equal(t, "cl-1", payload.CueListId)
		assert.Equal(t, "cue-1", payload.CueId)
		assert.Equal(t, "act-1", payload.ActionId)
	})

	t.Run("ActionDeleted returns correct subject and payload", func(t *testing.T) {
		ev := util.Event{
			Operation: model.OperationDeleted,
			EventData: map[string]string{
				model.MetadataCueListId: "cl-1",
				model.MetadataCueId:     "cue-1",
				model.MetadataActionId:  "act-1",
			},
		}
		sub, payload := api.ActionChanged(ev)
		assert.Equal(t, ActionDeletedEventSubject, sub)
		assert.Equal(t, "cl-1", payload.CueListId)
		assert.Equal(t, "cue-1", payload.CueId)
		assert.Equal(t, "act-1", payload.ActionId)
	})
}

func TestRegisterActionEvents(t *testing.T) {
	s, nc := testServer()
	defer nc.Close()
	defer s.Shutdown()

	m, err := model.NewCueingModel()
	require.NoError(t, err)

	messenger := messaging.NewMessenger(&messaging.MessengerCfg{Conn: nc})
	_, err = NewCueingApi(m, nil, messenger, nil)
	require.NoError(t, err)

	// Seed data
	clId, _, _ := m.CreateCueList(1, types.CueListTypeSequential)
	cueId, _, _ := m.CreateCue(clId, 0)

	t.Run("ActionCreated event is published", func(t *testing.T) {
		// Register a template first
		_ = m.RegisterActionTemplate(&types.ActionTemplate{TemplateId: "test-template"})

		wg := sync.WaitGroup{}
		wg.Add(1)
		var receivedId string

		sub, err := nc.Subscribe(ActionCreatedEventSubject, func(m *nats.Msg) {
			var event ActionChangeEvent
			_ = msgpack.Unmarshal(m.Data, &event)
			receivedId = event.ActionId
			wg.Done()
		})
		require.NoError(t, err)
		defer sub.Unsubscribe()

		actId, _, err := m.CreateAction(cueId, "test-template", 1)
		require.NoError(t, err)

		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			assert.Equal(t, actId, receivedId)
		case <-time.After(time.Second):
			t.Fatal("timed out waiting for ActionCreated event")
		}
	})

	t.Run("ActionUpdated event is published", func(t *testing.T) {
		actId, _, _ := m.CreateAction(cueId, "test-template", 2)

		wg := sync.WaitGroup{}
		wg.Add(1)
		var receivedId string

		sub, err := nc.Subscribe(ActionAttributesUpdatedEventSubject, func(m *nats.Msg) {
			var event ActionChangeEvent
			_ = msgpack.Unmarshal(m.Data, &event)
			receivedId = event.ActionId
			wg.Done()
		})
		require.NoError(t, err)
		defer sub.Unsubscribe()

		err = m.UpdateAction(actId, "label", "new label")
		require.NoError(t, err)

		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			assert.Equal(t, actId, receivedId)
		case <-time.After(time.Second):
			t.Fatal("timed out waiting for ActionUpdated event")
		}
	})

	t.Run("ActionDeleted event is published", func(t *testing.T) {
		actId, _, _ := m.CreateAction(cueId, "test-template", 3)

		wg := sync.WaitGroup{}
		wg.Add(1)
		var receivedId string

		sub, err := nc.Subscribe(ActionDeletedEventSubject, func(m *nats.Msg) {
			var event ActionChangeEvent
			_ = msgpack.Unmarshal(m.Data, &event)
			receivedId = event.ActionId
			wg.Done()
		})
		require.NoError(t, err)
		defer sub.Unsubscribe()

		err = m.DeleteAction(actId)
		require.NoError(t, err)

		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			assert.Equal(t, actId, receivedId)
		case <-time.After(time.Second):
			t.Fatal("timed out waiting for ActionDeleted event")
		}
	})
}
