// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package model

import (
	"testing"

	"github.com/stexxo/dynocue/components/cues/types"
	"github.com/stretchr/testify/assert"
)

func TestStartCueExecution(t *testing.T) {
	t.Parallel()

	t.Run("Success - New Execution", func(t *testing.T) {
		m, _ := NewCueingModel()
		clId, _, _ := m.CreateCueList(1, types.CueListTypeSequential)
		cueId, _, _ := m.CreateCue(clId, 0)
		err := m.StartCueExecution(cueId, true, true)
		assert.NoError(t, err)

		// Check that the entry exists
		selected, err := m.GetSelectedCue(clId)
		assert.NoError(t, err)
		assert.Equal(t, cueId, selected.CueId)
	})

	t.Run("Success - Change Selected Cue when first cue is still active", func(t *testing.T) {
		m, _ := NewCueingModel()
		clId, _, _ := m.CreateCueList(1, types.CueListTypeSequential)
		cueId, _, _ := m.CreateCue(clId, 0)
		cueId2, _, _ := m.CreateCue(clId, 0)
		err := m.StartCueExecution(cueId, true, true)
		assert.NoError(t, err)

		// Check that Cue 1 is selected
		selected, err := m.GetSelectedCue(clId)
		assert.NoError(t, err)
		assert.Equal(t, cueId, selected.CueId)

		// Trigger Execution of Cue 2
		err = m.StartCueExecution(cueId2, true, true)
		assert.NoError(t, err)

		// Check that Selected Cue is now Cue 2
		selected, err = m.GetSelectedCue(clId)
		assert.NoError(t, err)
		assert.Equal(t, cueId2, selected.CueId)
	})

	t.Run("Success - Change Selected Cue when first cue is not active", func(t *testing.T) {
		m, _ := NewCueingModel()
		clId, _, _ := m.CreateCueList(1, types.CueListTypeSequential)
		cueId, _, _ := m.CreateCue(clId, 0)
		cueId2, _, _ := m.CreateCue(clId, 0)
		err := m.StartCueExecution(cueId, true, false)
		assert.NoError(t, err)

		// Check that Cue 1 is selected
		selected, err := m.GetSelectedCue(clId)
		assert.NoError(t, err)
		assert.Equal(t, cueId, selected.CueId)

		// Trigger Execution of Cue 2
		err = m.StartCueExecution(cueId2, true, true)
		assert.NoError(t, err)

		// Check that Selected Cue is now Cue 2
		selected, err = m.GetSelectedCue(clId)
		assert.NoError(t, err)
		assert.Equal(t, cueId2, selected.CueId)
	})

	t.Run("cue doesnt exist", func(t *testing.T) {
		m, _ := NewCueingModel()
		err := m.StartCueExecution("non-existent", true, true)
		assert.ErrorIs(t, err, ErrCueNotFound)
	})
}

func BenchmarkCueStartExecution(b *testing.B) {
	m, _ := NewCueingModel()
	clId, _, _ := m.CreateCueList(1, types.CueListTypeSequential)
	cueId, _, _ := m.CreateCue(clId, 0)

	for b.Loop() {
		m.StartCueExecution(cueId, true, true)
	}
}

func BenchmarkCueStartExecutionFromPrevious(b *testing.B) {
	m, _ := NewCueingModel()
	clId, _, _ := m.CreateCueList(1, types.CueListTypeSequential)
	cueId1, _, _ := m.CreateCue(clId, 0)
	cueId2, _, _ := m.CreateCue(clId, 0)
	for b.Loop() {
		m.StartCueExecution(cueId1, true, true)

		b.StopTimer()
		m.StartCueExecution(cueId2, true, true)
		b.StartTimer()
	}
}

func TestCueStopExecution(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m, _ := NewCueingModel()
		clId, _, _ := m.CreateCueList(1, types.CueListTypeSequential)
		cueId, _, _ := m.CreateCue(clId, 0)
		err := m.StartCueExecution(cueId, true, true)
		assert.NoError(t, err)

		err = m.StopCueExecution(cueId)
		assert.NoError(t, err)

		// Check that cue is still selected
		selected, err := m.GetSelectedCue(clId)
		assert.NoError(t, err)
		assert.Equal(t, cueId, selected.CueId)
	})

	t.Run("Cue Finished after selection taken", func(t *testing.T) {
		m, _ := NewCueingModel()
		clId, _, _ := m.CreateCueList(1, types.CueListTypeSequential)
		cueId1, _, _ := m.CreateCue(clId, 0)
		cueId2, _, _ := m.CreateCue(clId, 0)

		// Start Cue 1
		err := m.StartCueExecution(cueId1, true, true)
		assert.NoError(t, err)

		// check that it is selected
		selected, err := m.GetSelectedCue(clId)
		assert.NoError(t, err)
		assert.Equal(t, cueId1, selected.CueId)

		// Start Cue 2
		err = m.StartCueExecution(cueId2, true, true)
		assert.NoError(t, err)

		// Check that cue 2 is selected
		selected, err = m.GetSelectedCue(clId)
		assert.NoError(t, err)
		assert.Equal(t, cueId2, selected.CueId)

		// Stop Cue 1
		err = m.StopCueExecution(cueId1)
		assert.NoError(t, err)

		// Check that cue 2 is still selected
		selected, err = m.GetSelectedCue(clId)
		assert.NoError(t, err)
		assert.Equal(t, cueId2, selected.CueId)

		// Check that Cue 1 is no longer found in the execution records
		_, err = m.GetCueExecution(cueId1)
		assert.ErrorIs(t, err, ErrCueNotFound)

		// Idempotent Stop
		err = m.StopCueExecution(cueId2)
		assert.NoError(t, err)
	})
}

func BenchmarkCueStopExecutionStaySelected(b *testing.B) {
	m, _ := NewCueingModel()
	clId, _, _ := m.CreateCueList(1, types.CueListTypeSequential)
	cueId, _, _ := m.CreateCue(clId, 0)
	m.StartCueExecution(cueId, true, true)

	for b.Loop() {
		m.StopCueExecution(cueId)
		b.StopTimer()
		m.StartCueExecution(cueId, true, true)
		b.StartTimer()
	}
}

func BenchmarkCueStopExecutionChangeSelected(b *testing.B) {
	m, _ := NewCueingModel()
	clId, _, _ := m.CreateCueList(1, types.CueListTypeSequential)
	cueId1, _, _ := m.CreateCue(clId, 0)
	m.StartCueExecution(cueId1, false, true)

	for b.Loop() {
		m.StopCueExecution(cueId1)
		b.StopTimer()
		m.StartCueExecution(cueId1, false, true)
		b.StartTimer()
	}
}
