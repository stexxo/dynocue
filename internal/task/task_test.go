// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package engine

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTaskEngine(t *testing.T) {
	t.Parallel()

	t.Run("Initialization", func(t *testing.T) {
		t.Parallel()
		interval := 15 * time.Millisecond
		e := NewEngine(interval)

		assert.Equal(t, interval, e.tickInterval)
		assert.NotNil(t, e.incoming)
		assert.False(t, e.isRunning)
	})
}

func TestTaskEngine_Start(t *testing.T) {
	t.Parallel()

	t.Run("Normal Start", func(t *testing.T) {
		t.Parallel()
		e := NewEngine(10 * time.Millisecond)
		defer e.Stop()

		e.Start()
		assert.True(t, e.isRunning)
	})

	// Verifies that multiple Start calls do not disrupt engine state or context.
	t.Run("Idempotency", func(t *testing.T) {
		t.Parallel()
		e := NewEngine(10 * time.Millisecond)
		e.Start()
		defer e.Stop()

		oldCtx := e.ctx
		assert.NotPanics(t, func() { e.Start() })
		assert.Equal(t, oldCtx, e.ctx)
	})
}

func TestTaskEngine_Stop(t *testing.T) {
	t.Parallel()

	t.Run("Stop Running Engine", func(t *testing.T) {
		t.Parallel()
		e := NewEngine(10 * time.Millisecond)
		e.Start()
		e.Stop()
		assert.False(t, e.isRunning)
		assert.Error(t, e.ctx.Err(), "Context should be cancelled upon engine stop")
	})

	t.Run("Stop Unstarted Engine", func(t *testing.T) {
		t.Parallel()
		e := NewEngine(10 * time.Millisecond)
		assert.NotPanics(t, func() { e.Stop() })
	})
}

func TestTaskEngine_AddTask(t *testing.T) {
	t.Parallel()

	t.Run("Add Valid Task", func(t *testing.T) {
		t.Parallel()
		e := NewEngine(10 * time.Millisecond)

		e.AddTask(func(dt time.Duration) bool { return true })
		assert.Equal(t, 1, len(e.incoming))
	})

	// Ensures the engine gracefully ignores nil function pointers.
	t.Run("Nil Task Safety", func(t *testing.T) {
		t.Parallel()
		e := NewEngine(10 * time.Millisecond)

		e.AddTask(nil)
		assert.Equal(t, 0, len(e.incoming))
	})

	// Verifies that AddTask defaults to a non-blocking drop strategy when the channel is saturated.
	t.Run("Failure Mode: Buffer Saturation", func(t *testing.T) {
		t.Parallel()
		e := NewEngine(1 * time.Second)

		for i := 0; i < 256; i++ {
			e.AddTask(func(dt time.Duration) bool { return false })
		}

		start := time.Now()
		e.AddTask(func(dt time.Duration) bool { return false })

		assert.WithinDuration(t, time.Now(), start, 20*time.Millisecond, "Operation must not block caller")
	})
}

func TestTaskEngine_ProcessLoop(t *testing.T) {
	t.Parallel()

	// Validates that tasks returning 'true' are successfully decommissioned from the active set.
	t.Run("Task Execution and Removal", func(t *testing.T) {
		t.Parallel()
		e := NewEngine(5 * time.Millisecond)
		e.Start()
		defer e.Stop()

		executed := make(chan struct{})
		e.AddTask(func(dt time.Duration) bool {
			close(executed)
			return true
		})

		select {
		case <-executed:
			require.Eventually(t, func() bool {
				return len(e.activeTasks) == 0
			}, 100*time.Millisecond, 5*time.Millisecond)
		case <-time.After(200 * time.Millisecond):
			t.Fatal("Task execution timeout")
		}
	})
}

func TestTaskEngine_RunTaskSafely(t *testing.T) {
	t.Parallel()

	// Ensures that a panic within a TaskFn is recovered and does not terminate the engine loop.
	t.Run("Failure Mode: Panic Recovery", func(t *testing.T) {
		t.Parallel()
		e := NewEngine(5 * time.Millisecond)
		e.Start()
		defer e.Stop()

		canary := int32(0)
		e.AddTask(func(dt time.Duration) bool {
			panic("simulated task panic")
		})

		e.AddTask(func(dt time.Duration) bool {
			atomic.StoreInt32(&canary, 1)
			return false
		})

		require.Eventually(t, func() bool {
			return atomic.LoadInt32(&canary) == 1
		}, 100*time.Millisecond, 5*time.Millisecond, "Sibling tasks must continue to execute after panic")
	})
}
