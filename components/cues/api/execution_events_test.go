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

func TestExecutionEventsMethods(t *testing.T) {
	api := &CueingApi{}

	t.Run("ExecutionStarted returns correct subject and payload", func(t *testing.T) {
		ev := util.Event{
			Operation: model.OperationStarted,
			EventData: map[string]string{
				model.MetadataCueListId: "cl-1",
				model.MetadataCueId:     "cue-1",
			},
		}
		sub, payload := api.ExecutionChanged(ev)
		assert.Equal(t, ExecutionStartedEventSubject, sub)
		assert.Equal(t, "cl-1", payload.CueListId)
		assert.Equal(t, "cue-1", payload.CueId)
	})

	t.Run("ExecutionFinished returns correct subject and payload", func(t *testing.T) {
		ev := util.Event{
			Operation: model.OperationFinished,
			EventData: map[string]string{
				model.MetadataCueListId: "cl-1",
				model.MetadataCueId:     "cue-1",
			},
		}
		sub, _ := api.ExecutionChanged(ev)
		assert.Equal(t, ExecutionFinishedEventSubject, sub)
	})

	t.Run("ExecutionUnselected returns correct subject and payload", func(t *testing.T) {
		ev := util.Event{
			Operation: model.OperationUnselected,
			EventData: map[string]string{
				model.MetadataCueListId: "cl-1",
				model.MetadataCueId:     "cue-1",
			},
		}
		sub, _ := api.ExecutionChanged(ev)
		assert.Equal(t, ExecutionUnselectedEventSubject, sub)
	})

	t.Run("ExecutionDeleted returns correct subject and payload", func(t *testing.T) {
		ev := util.Event{
			Operation: model.OperationDeleted,
			EventData: map[string]string{
				model.MetadataCueListId: "cl-1",
				model.MetadataCueId:     "cue-1",
			},
		}
		sub, _ := api.ExecutionChanged(ev)
		assert.Equal(t, ExecutionDeletedEventSubject, sub)
	})
}

func TestRegisterExecutionEvents(t *testing.T) {
	setup := func(t *testing.T) (*model.CueingModel, *nats.Conn, func()) {
		s, nc := testServer()
		m, err := model.NewCueingModel()
		require.NoError(t, err)

		messenger := messaging.NewMessenger(&messaging.MessengerCfg{Conn: nc})
		_, err = NewCueingApi(m, nil, nil, messenger, nil)
		require.NoError(t, err)

		return m, nc, func() {
			nc.Close()
			s.Shutdown()
		}
	}

	t.Run("ExecutionStarted event is published", func(t *testing.T) {
		m, nc, cleanup := setup(t)
		defer cleanup()

		// Seed data
		clId, _, _ := m.CreateCueList(1, types.CueListTypeSequential)
		cueId, _, _ := m.CreateCue(clId, 1)

		wg := sync.WaitGroup{}
		wg.Add(1)
		var receivedId string

		sub, err := nc.Subscribe(ExecutionStartedEventSubject, func(m *nats.Msg) {
			var event ExecutionChangeEvent
			_ = msgpack.Unmarshal(m.Data, &event)
			receivedId = event.CueId
			wg.Done()
		})
		require.NoError(t, err)
		defer sub.Unsubscribe()

		err = m.StartCueExecution(cueId, true, true)
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
			t.Fatal("timed out waiting for ExecutionStarted event")
		}
	})

	t.Run("ExecutionFinished event is published", func(t *testing.T) {
		m, nc, cleanup := setup(t)
		defer cleanup()

		// Seed data
		clId, _, _ := m.CreateCueList(1, types.CueListTypeSequential)
		cueId, _, _ := m.CreateCue(clId, 1)

		wg := sync.WaitGroup{}
		wg.Add(1)
		var receivedId string

		sub, err := nc.Subscribe(ExecutionFinishedEventSubject, func(m *nats.Msg) {
			var event ExecutionChangeEvent
			_ = msgpack.Unmarshal(m.Data, &event)
			receivedId = event.CueId
			wg.Done()
		})
		require.NoError(t, err)
		defer sub.Unsubscribe()

		_ = m.StartCueExecution(cueId, true, true)
		err = m.StopCueExecution(cueId)
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
			t.Fatal("timed out waiting for ExecutionFinished event")
		}
	})

	t.Run("ExecutionUnselected event is published", func(t *testing.T) {
		m, nc, cleanup := setup(t)
		defer cleanup()

		// Seed data
		clId, _, _ := m.CreateCueList(1, types.CueListTypeSequential)
		cueId, _, _ := m.CreateCue(clId, 1)
		cueId2, _, _ := m.CreateCue(clId, 2)

		// Start first cue as selected and active
		_ = m.StartCueExecution(cueId, true, true)

		wg := sync.WaitGroup{}
		wg.Add(1)
		var receivedId string

		sub, err := nc.Subscribe(ExecutionUnselectedEventSubject, func(m *nats.Msg) {
			var event ExecutionChangeEvent
			_ = msgpack.Unmarshal(m.Data, &event)
			receivedId = event.CueId
			wg.Done()
		})
		require.NoError(t, err)
		defer sub.Unsubscribe()

		// Start second cue as selected, will unselect first cue (because it's active)
		err = m.StartCueExecution(cueId2, true, false)
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
			t.Fatal("timed out waiting for ExecutionUnselected event")
		}
	})

	t.Run("ExecutionDeleted event is published", func(t *testing.T) {
		m, nc, cleanup := setup(t)
		defer cleanup()

		// Seed data
		clId, _, _ := m.CreateCueList(1, types.CueListTypeSequential)
		cueId3, _, _ := m.CreateCue(clId, 3)

		// Start third cue as selected (it's not active)
		_ = m.StartCueExecution(cueId3, true, false)

		wg := sync.WaitGroup{}
		wg.Add(1)
		var receivedId string

		sub, err := nc.Subscribe(ExecutionDeletedEventSubject, func(m *nats.Msg) {
			var event ExecutionChangeEvent
			_ = msgpack.Unmarshal(m.Data, &event)
			receivedId = event.CueId
			wg.Done()
		})
		require.NoError(t, err)
		defer sub.Unsubscribe()

		// Start another cue as selected, will delete cueId3 from execution because it's not active
		cueId4, _, _ := m.CreateCue(clId, 4)
		err = m.StartCueExecution(cueId4, true, false)
		require.NoError(t, err)

		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			assert.Equal(t, cueId3, receivedId)
		case <-time.After(time.Second):
			t.Fatal("timed out waiting for ExecutionDeleted event")
		}
	})
}
