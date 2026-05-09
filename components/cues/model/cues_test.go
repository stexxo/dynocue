package model

import (
	"testing"

	"github.com/stexxo/dynocue/components/cues/types"
	"github.com/stretchr/testify/assert"
)

func TestCreateCue(t *testing.T) {
	t.Run("CueList Doesnt exist", func(t *testing.T) {
		m, _ := NewCueingModel()
		id, num, err := m.CreateCue("notreal", 10)
		assert.Empty(t, id)
		assert.Equal(t, uint(0), num)
		assert.ErrorIs(t, err, ErrCueListNotFound)
	})

	t.Run("CueList exists, but no cues exist", func(t *testing.T) {
		m, _ := NewCueingModel()
		cueListId, _, _ := m.CreateCueList(1, types.CueListTypeSequential)
		id, num, err := m.CreateCue(cueListId, 0)
		assert.NotEmpty(t, id)
		assert.Equal(t, uint(1), num)
		assert.NoError(t, err)
	})

	t.Run("CueList exists, and cues exist", func(t *testing.T) {
		m, _ := NewCueingModel()
		cueListId, _, _ := m.CreateCueList(1, types.CueListTypeSequential)
		_, num, _ := m.CreateCue(cueListId, 0)
		assert.Equal(t, uint(1), num)
		id, num, err := m.CreateCue(cueListId, 0)
		assert.NotEmpty(t, id)
		assert.Equal(t, uint(2), num)
		assert.NoError(t, err)
	})

	t.Run("create with specified number", func(t *testing.T) {
		m, _ := NewCueingModel()
		cueListId, _, _ := m.CreateCueList(1, types.CueListTypeSequential)
		id, num, err := m.CreateCue(cueListId, 10)
		assert.NotEmpty(t, id)
		assert.Equal(t, uint(10), num)
		assert.NoError(t, err)
	})

	t.Run("create with specified number with conflict", func(t *testing.T) {
		m, _ := NewCueingModel()
		cueListId, _, _ := m.CreateCueList(1, types.CueListTypeSequential)
		id, num, err := m.CreateCue(cueListId, 10)
		assert.NotEmpty(t, id)
		assert.Equal(t, uint(10), num)
		assert.NoError(t, err)

		id, num, err = m.CreateCue(cueListId, 10)
		assert.Empty(t, id)
		assert.Equal(t, uint(0), num)
		assert.ErrorIs(t, err, ErrNumberExists)
	})
}

func TestEnumerateCues(t *testing.T) {
	t.Run("Get all", func(t *testing.T) {
		m, _ := NewCueingModel()
		cueListId, _, _ := m.CreateCueList(1, types.CueListTypeSequential)
		id, num, err := m.CreateCue(cueListId, 0)
		assert.NotEmpty(t, id)
		assert.Equal(t, uint(1), num)
		assert.NoError(t, err)

		cues, err := m.EnumerateCues(cueListId)
		assert.NoError(t, err)
		assert.Len(t, cues, 1)
	})
}

func TestGetCueByNumber(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m, _ := NewCueingModel()
		cueListId, _, _ := m.CreateCueList(1, types.CueListTypeSequential)
		id, num, err := m.CreateCue(cueListId, 0)
		assert.NotEmpty(t, id)
		assert.Equal(t, uint(1), num)
		assert.NoError(t, err)

		cue, err := m.GetCueByNumber(cueListId, num)
		assert.NoError(t, err)
		assert.Equal(t, id, cue.CueId)
	})

	t.Run("Cue Not Found", func(t *testing.T) {
		m, _ := NewCueingModel()
		cueListId, _, _ := m.CreateCueList(1, types.CueListTypeSequential)

		_, err := m.GetCueByNumber(cueListId, 1)
		assert.ErrorIs(t, err, ErrCueNotFound)
	})

	t.Run("Cues in list and cue not found", func(t *testing.T) {
		m, _ := NewCueingModel()
		cueListId, _, _ := m.CreateCueList(1, types.CueListTypeSequential)
		id, _, err := m.CreateCue(cueListId, 10)
		assert.NotEmpty(t, id)
		assert.NoError(t, err)

		_, err = m.GetCueByNumber(cueListId, 1)
		assert.ErrorIs(t, err, ErrCueNotFound)
	})
}

func TestGetCueById(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m, _ := NewCueingModel()
		cueListId, _, _ := m.CreateCueList(1, types.CueListTypeSequential)
		id, _, err := m.CreateCue(cueListId, 0)
		assert.NotEmpty(t, id)
		assert.NoError(t, err)

		cue, err := m.GetCueById(id)
		assert.NoError(t, err)
		assert.Equal(t, id, cue.CueId)
	})

	t.Run("Cue Not Found", func(t *testing.T) {
		m, _ := NewCueingModel()
		_, err := m.GetCueById("notreal")
		assert.ErrorIs(t, err, ErrCueNotFound)
	})
}

func TestDeleteCueById(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m, _ := NewCueingModel()
		cueListId, _, _ := m.CreateCueList(1, types.CueListTypeSequential)
		id, _, err := m.CreateCue(cueListId, 0)
		assert.NotEmpty(t, id)
		assert.NoError(t, err)

		err = m.DeleteCueById(id)
		assert.NoError(t, err)

		t.Run("Idempotent", func(t *testing.T) {
			err = m.DeleteCueById(id)
			assert.NoError(t, err)
		})
	})
}

func TestDeleteAllCuesByCueListId(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m, _ := NewCueingModel()
		cueListId, _, _ := m.CreateCueList(1, types.CueListTypeSequential)
		for range 10 {
			_, _, _ = m.CreateCue(cueListId, 0)
		}
		cues, err := m.EnumerateCues(cueListId)
		assert.NoError(t, err)
		assert.Len(t, cues, 10)

		err = m.DeleteAllCuesByCueListId(cueListId)
		assert.NoError(t, err)

		cues, err = m.EnumerateCues(cueListId)
		assert.NoError(t, err)
		assert.Len(t, cues, 0)
		t.Run("Idempotent", func(t *testing.T) {
			err = m.DeleteAllCuesByCueListId(cueListId)
			assert.NoError(t, err)
		})
	})
}

func TestUpdateCueAttribute(t *testing.T) {
	t.Run("Update Cue that doesn't exist", func(t *testing.T) {
		m, _ := NewCueingModel()
		err := m.UpdateCueAttribute("notreal", "label", "New Label")
		assert.ErrorIs(t, err, ErrCueNotFound)
	})

	t.Run("Update Cue that exists", func(t *testing.T) {
		m, _ := NewCueingModel()
		cueListId, _, _ := m.CreateCueList(1, types.CueListTypeSequential)
		id, _, _ := m.CreateCue(cueListId, 0)
		err := m.UpdateCueAttribute(id, "label", "New Label")
		assert.NoError(t, err)
	})
}
