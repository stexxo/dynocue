package model

import (
	"math/rand/v2"
	"runtime"
	"testing"

	"github.com/stexxo/dynocue/components/cues/types"
	"github.com/stretchr/testify/assert"
)

func TestCreateCueList(t *testing.T) {
	t.Parallel()

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
		assert.ErrorIs(t, err, ErrNumberExists)
	})
}

func TestEnumerateCueLists(t *testing.T) {
	t.Parallel()

	m, _ := NewCueingModel()
	_, _, _ = m.CreateCueList(1, types.CueListTypeSequential)
	_, _, _ = m.CreateCueList(2, types.CueListTypeSequential)

	lists, err := m.EnumerateCueLists()
	assert.NoError(t, err)
	assert.Len(t, lists, 2)
	assert.Equal(t, uint(1), lists[0].Number)
	assert.Equal(t, uint(2), lists[1].Number)
}

func TestGetCueListByNumber(t *testing.T) {
	t.Parallel()

	m, _ := NewCueingModel()
	_, _, _ = m.CreateCueList(10, types.CueListTypeSequential)

	t.Run("Found", func(t *testing.T) {
		cl, err := m.GetCueListByNumber(10)
		assert.NoError(t, err)
		assert.Equal(t, uint(10), cl.Number)
	})

	t.Run("NotFound", func(t *testing.T) {
		cl, err := m.GetCueListByNumber(20)
		assert.ErrorIs(t, err, ErrCueListNotFound)
		assert.Nil(t, cl)
	})
}

func TestGetCueListById(t *testing.T) {
	t.Parallel()

	m, _ := NewCueingModel()
	id, _, _ := m.CreateCueList(10, types.CueListTypeSequential)

	t.Run("Found", func(t *testing.T) {
		cl, err := m.GetCueListById(id)
		assert.NoError(t, err)
		assert.Equal(t, id, cl.CueListId)
	})

	t.Run("NotFound", func(t *testing.T) {
		cl, err := m.GetCueListById("non-existent")
		assert.ErrorIs(t, err, ErrCueListNotFound)
		assert.Nil(t, cl)
	})
}

func TestDeleteCueListById(t *testing.T) {
	t.Parallel()

	m, _ := NewCueingModel()
	id, _, _ := m.CreateCueList(10, types.CueListTypeSequential)

	err := m.DeleteCueListById(id)
	assert.NoError(t, err)

	_, err = m.GetCueListById(id)
	assert.ErrorIs(t, err, ErrCueListNotFound)
}

func TestUpdateCueListAttribute(t *testing.T) {
	t.Parallel()

	t.Run("Update CueList that doesn't exist", func(t *testing.T) {
		m, _ := NewCueingModel()
		err := m.UpdateCueListAttribute("notreal", "label", "New Label")
		assert.ErrorIs(t, err, ErrCueListNotFound)
	})

	t.Run("Update Cue that exists", func(t *testing.T) {
		m, _ := NewCueingModel()
		id, _, _ := m.CreateCueList(10, types.CueListTypeSequential)

		err := m.UpdateCueListAttribute(id, "label", "New Label")
		assert.NoError(t, err)

		cl, _ := m.GetCueListById(id)
		assert.Equal(t, "New Label", cl.Label)
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

func BenchmarkGetCueListById(b *testing.B) {
	m, _ := NewCueingModel()
	var targetId string
	for i := range 1000 {
		id, _, _ := m.CreateCueList(uint(i+1), types.CueListTypeSequential)
		if i == 500 {
			targetId = id
		}
	}
	b.ResetTimer()
	for b.Loop() {
		_, _ = m.GetCueListById(targetId)
	}
}

func BenchmarkCreateAndGetParallel(b *testing.B) {
	m, _ := NewCueingModel()
	var targetId string
	for i := range 1000 {
		id, _, _ := m.CreateCueList(uint(i+1), types.CueListTypeSequential)
		if i == 500 {
			targetId = id
		}
	}
	b.SetParallelism(runtime.NumCPU())
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			if i%2 == 0 {
				_, _, _ = m.CreateCueList(0, types.CueListTypeSequential)
			} else {
				_, _ = m.GetCueListById(targetId)
			}
			i++
		}
	})
}

func BenchmarkEnumerateCueLists(b *testing.B) {
	m, _ := NewCueingModel()
	for i := range 1000 {
		_, _, _ = m.CreateCueList(uint(i+1), types.CueListTypeSequential)
	}
	b.ResetTimer()
	for b.Loop() {
		_, _ = m.EnumerateCueLists()
	}
}

func BenchmarkGetCueListByNumber(b *testing.B) {
	m, _ := NewCueingModel()
	for i := range 1000 {
		_, _, _ = m.CreateCueList(uint(i+1), types.CueListTypeSequential)
	}
	b.ResetTimer()
	for b.Loop() {
		target := uint(rand.IntN(1000) + 1)
		_, _ = m.GetCueListByNumber(target)
	}
}

func BenchmarkDeleteCueListById(b *testing.B) {
	m, _ := NewCueingModel()
	var ids []string
	for i := range 100 {
		id, _, _ := m.CreateCueList(uint(i+1), types.CueListTypeSequential)
		ids = append(ids, id)
	}

	b.ResetTimer()
	for b.Loop() {
		idx := rand.IntN(len(ids))
		targetId := ids[idx]

		_ = m.DeleteCueListById(targetId)

		b.StopTimer()
		// Recreate the deleted one with the same ID logic is not possible easily
		// because CreateCueList generates a new UUID.
		// We just need to replace the ID in our tracking slice.
		newId, _, _ := m.CreateCueList(0, types.CueListTypeSequential)
		ids[idx] = newId
		b.StartTimer()
	}
}

func BenchmarkUpdateCueListAttribute(b *testing.B) {
	m, _ := NewCueingModel()
	var targetId string
	for i := range 1000 {
		id, _, _ := m.CreateCueList(uint(i+1), types.CueListTypeSequential)
		if i == 500 {
			targetId = id
		}
	}
	b.ResetTimer()
	for b.Loop() {
		_ = m.UpdateCueListAttribute(targetId, "label", "Updated Label")
	}
}
