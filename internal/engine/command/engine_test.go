package engine

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommandEngine_Lifecycle(t *testing.T) {
	t.Parallel()

	t.Run("Idempotent Operations", func(t *testing.T) {
		t.Parallel()
		e := NewCommandEngine(10)

		// Verify multiple starts do not cause issues
		assert.NotPanics(t, func() {
			e.Start()
			e.Start()
		})
		assert.True(t, e.running)

		// Verify multiple stops do not cause issues
		assert.NotPanics(t, func() {
			e.Stop()
			e.Stop()
		})
		assert.False(t, e.running)
	})
}

func TestCommandEngine_Execution(t *testing.T) {
	t.Parallel()

	t.Run("Serialized FIFO Execution", func(t *testing.T) {
		t.Parallel()
		e := NewCommandEngine(10)
		e.Start()
		defer e.Stop()

		var results []int
		done := make(chan struct{})

		// We add commands that must execute in order
		for i := 1; i <= 3; i++ {
			val := i
			e.AddCommand(func() {
				results = append(results, val)
				if val == 3 {
					close(done)
				}
			})
		}

		select {
		case <-done:
			assert.Equal(t, []int{1, 2, 3}, results, "Commands must execute in the order they were added")
		case <-time.After(500 * time.Millisecond):
			t.Fatal("Commands timed out")
		}
	})

	t.Run("Panic Recovery", func(t *testing.T) {
		t.Parallel()
		e := NewCommandEngine(10)
		e.Start()
		defer e.Stop()

		var completed atomic.Bool

		// This command will panic
		e.AddCommand(func() {
			panic("critical failure")
		})

		// This command should still run because the engine recovered
		e.AddCommand(func() {
			completed.Store(true)
		})

		require.Eventually(t, func() bool {
			return completed.Load()
		}, 200*time.Millisecond, 10*time.Millisecond, "Engine should process subsequent commands after a panic")
	})
}

func TestCommandEngine_AddCommand(t *testing.T) {
	t.Parallel()

	t.Run("Nil Command Safety", func(t *testing.T) {
		t.Parallel()
		e := NewCommandEngine(10)
		e.Start()
		defer e.Stop()

		assert.NotPanics(t, func() {
			e.AddCommand(nil)
		})
	})

	t.Run("Non-Blocking On Full Buffer", func(t *testing.T) {
		t.Parallel()
		// Engine with size 1, but we don't Start it so the buffer stays full
		e := NewCommandEngine(1)

		e.mu.Lock()
		e.running = true // Fake running state to pass the check
		e.mu.Unlock()

		// Fill the one slot
		e.AddCommand(func() {})

		// The second add should hit the 'default' case and return immediately
		start := time.Now()
		e.AddCommand(func() {})

		assert.WithinDuration(t, time.Now(), start, 20*time.Millisecond, "AddCommand must not block when buffer is full")
	})
}
