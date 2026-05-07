package model

import (
	"runtime"
	"testing"

	"github.com/stexxo/dynocue/components/cues/types"
	"github.com/stretchr/testify/assert"
)

func TestCreateCueList(t *testing.T) {
	t.Run("Create with 0 number with no other cuelists", func(t *testing.T) {
		m, _ := NewCueingModel()
		id, number, err := m.CreateCueList(0, types.CueListTypeSequential)
		assert.NoError(t, err)
		assert.Equal(t, uint(1), number)
		assert.NotEmpty(t, id)
	})

	t.Run("Create with 0 number with 1 other cuelist", func(t *testing.T) {
		m, _ := NewCueingModel()
		id, number, err := m.CreateCueList(0, types.CueListTypeSequential)
		assert.NoError(t, err)
		assert.Equal(t, uint(1), number)
		assert.NotEmpty(t, id)

		id, number, err = m.CreateCueList(0, types.CueListTypeSequential)
		assert.NoError(t, err)
		assert.Equal(t, uint(2), number)
		assert.NotEmpty(t, id)
	})

	t.Run("Create with specified number with no conflict", func(t *testing.T) {
		m, _ := NewCueingModel()
		id, number, err := m.CreateCueList(10, types.CueListTypeSequential)
		assert.NoError(t, err)
		assert.Equal(t, uint(10), number)
		assert.NotEmpty(t, id)
	})

	t.Run("Create with specified number with conflict", func(t *testing.T) {
		m, _ := NewCueingModel()
		id, number, err := m.CreateCueList(10, types.CueListTypeSequential)
		assert.NoError(t, err)
		assert.Equal(t, uint(10), number)
		assert.NotEmpty(t, id)

		id, number, err = m.CreateCueList(10, types.CueListTypeSequential)
		assert.ErrorIs(t, err, ErrCueListExists)
	})
}

func BenchmarkCreateCueList(b *testing.B) {
	m, _ := NewCueingModel()
	for b.Loop() {
		m.CreateCueList(0, types.CueListTypeSequential)
	}
}

func BenchmarkCreateCueListCreateParallel(b *testing.B) {
	m, _ := NewCueingModel()
	b.SetParallelism(runtime.NumCPU())
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.CreateCueList(0, types.CueListTypeSequential)
		}
	})
}
