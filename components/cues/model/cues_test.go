package model

import (
	"github.com/stexxo/dynocue/components/cues/types"
	"github.com/stretchr/testify/assert"
	"testing"
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
