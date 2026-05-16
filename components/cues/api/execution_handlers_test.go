// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package api

import (
	"errors"
	"testing"

	"github.com/stexxo/dynocue/components/cues/engine"
	"github.com/stexxo/dynocue/components/cues/model"
	"github.com/stexxo/dynocue/components/cues/types"
	"github.com/stexxo/dynocue/core/messaging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupExecution(t *testing.T) (*model.CueingModel, *engine.CueingEngine, *CueingApi) {
	m, err := model.NewCueingModel()
	require.NoError(t, err)
	eng := engine.NewCueingEngine(m, nil, nil)
	api := &CueingApi{
		model:  m,
		engine: eng,
	}
	return m, eng, api
}

func TestGoToCue(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m, _, api := setupExecution(t)

		clId, _, _ := m.CreateCueList(1, types.CueListTypeSequential)
		cueId, _, _ := m.CreateCue(clId, 1)

		req := &GoToCueRequest{
			CueId: cueId,
		}

		resp, err := api.GoToCue("test-sub", req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)

		// Verify execution started in model
		exec, err := m.GetCueExecution(cueId)
		assert.NoError(t, err)
		assert.True(t, exec.Selected)
	})

	t.Run("Error not found", func(t *testing.T) {
		_, _, api := setupExecution(t)

		req := &GoToCueRequest{
			CueId: "non-existent",
		}

		resp, err := api.GoToCue("test-sub", req)
		assert.Error(t, err)
		assert.Nil(t, resp)

		assert.True(t, errors.Is(err, model.ErrCueNotFound))
		friendlyErr, ok := errors.AsType[*messaging.FriendlyError](err)
		assert.True(t, ok)
		assert.Equal(t, CueNotFound, friendlyErr.FriendlyErr)
	})
}

func TestGoToNextCue(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m, _, api := setupExecution(t)

		clId, _, _ := m.CreateCueList(1, types.CueListTypeSequential)
		cueId1, _, _ := m.CreateCue(clId, 1)
		cueId2, _, _ := m.CreateCue(clId, 2)

		// Select first cue
		err := m.StartCueExecution(cueId1, true, false)
		require.NoError(t, err)

		req := &GoToNextCueRequest{
			CueListId: clId,
		}

		resp, err := api.GoToNextCue("test-sub", req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)

		// Verify next cue is now selected
		exec1, err := m.GetCueExecution(cueId1)
		if err == nil {
			assert.False(t, exec1.Selected)
		} else {
			assert.ErrorIs(t, err, model.ErrCueNotFound)
		}

		exec2, err := m.GetCueExecution(cueId2)
		assert.NoError(t, err)
		assert.True(t, exec2.Selected)
	})

	t.Run("Error no cue selected", func(t *testing.T) {
		m, _, api := setupExecution(t)

		clId, _, _ := m.CreateCueList(1, types.CueListTypeSequential)
		_, _, _ = m.CreateCue(clId, 1)

		req := &GoToNextCueRequest{
			CueListId: clId,
		}

		resp, err := api.GoToNextCue("test-sub", req)
		assert.Error(t, err)
		assert.Nil(t, resp)

		friendlyErr, ok := errors.AsType[*messaging.FriendlyError](err)
		assert.True(t, ok)
		assert.Equal(t, NoCueSelected, friendlyErr.FriendlyErr)
	})

	t.Run("Error no next cue", func(t *testing.T) {
		m, _, api := setupExecution(t)

		clId, _, _ := m.CreateCueList(1, types.CueListTypeSequential)
		cueId, _, _ := m.CreateCue(clId, 1)

		// Select the only cue
		err := m.StartCueExecution(cueId, true, false)
		require.NoError(t, err)

		req := &GoToNextCueRequest{
			CueListId: clId,
		}

		resp, err := api.GoToNextCue("test-sub", req)
		assert.Error(t, err)
		assert.Nil(t, resp)

		assert.True(t, errors.Is(err, model.ErrNoNextCue))
		friendlyErr, ok := errors.AsType[*messaging.FriendlyError](err)
		assert.True(t, ok)
		assert.Equal(t, NoNextCue, friendlyErr.FriendlyErr)
	})
}
