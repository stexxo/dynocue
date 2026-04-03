// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package task

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// TaskFn represents a unit of work executed periodically by the engine.
// It receives the elapsed time since the last tick (dt) and returns true
// if the task has completed and should be decommissioned from the engine.
type TaskFn func(dt time.Duration) (finished bool)

// Engine manages the lifecycle and periodic execution of recurring tasks.
// It utilizes a double-buffered approach via a staged incoming channel and
// an active execution slice to ensure thread-safe task registration.
type Engine struct {
	tickInterval time.Duration
	activeTasks  []TaskFn
	incoming     chan TaskFn

	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
	mu        sync.Mutex
	isRunning bool
}

// NewEngine initializes a new engine with the specified tick frequency.
// The incoming task buffer is initialized with a capacity of 256 to minimize
// pressure on the registration call sites.
func NewEngine(interval time.Duration) *Engine {
	return &Engine{
		tickInterval: interval,
		incoming:     make(chan TaskFn, 256),
	}
}

// Start spawns the background worker goroutine responsible for task execution.
// This method is idempotent; calling it on an already running engine has no effect.
func (e *Engine) Start() {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.isRunning {
		return
	}

	e.ctx, e.cancel = context.WithCancel(context.Background())
	e.isRunning = true
	e.wg.Add(1)

	go e.processLoop()
	slog.Info("Engine: Started", "interval", e.tickInterval)
}

// Stop signals the execution loop to terminate and blocks until all cleanup
// operations are complete and the worker goroutine has exited.
func (e *Engine) Stop() {
	e.mu.Lock()
	if !e.isRunning {
		e.mu.Unlock()
		return
	}
	e.isRunning = false
	e.mu.Unlock()

	e.cancel()
	e.wg.Wait()
	slog.Info("Engine: Stopped")
}

// AddTask submits a new TaskFn to the engine. Tasks are staged in an
// internal buffer and integrated into the active execution set at the
// beginning of the next tick. If the internal buffer is full, the task
// is dropped and a warning is logged.
func (e *Engine) AddTask(t TaskFn) {
	if t == nil {
		return
	}
	select {
	case e.incoming <- t:
	default:
		slog.Warn("Engine: Queue full, dropping task")
	}
}

// processLoop handles the primary execution lifecycle, including ticker
// management, task migration from the staging channel, and in-place
// slice filtering for completed tasks.
func (e *Engine) processLoop() {
	defer e.wg.Done()
	ticker := time.NewTicker(e.tickInterval)
	defer ticker.Stop()

	last := time.Now()

	for {
		select {
		case <-e.ctx.Done():
			return
		case now := <-ticker.C:
			dt := now.Sub(last)
			last = now

			// Migrate staged tasks to the active list
			e.drainIncoming()

			// Update tasks and perform in-place deletion.
			// Reuses the underlying array of activeTasks to minimize allocations.
			n := 0
			for _, task := range e.activeTasks {
				if !e.runTaskSafely(task, dt) {
					e.activeTasks[n] = task
					n++
				}
			}

			// Zero out the remainder of the slice to prevent memory leaks
			// by allowing the GC to reclaim stale TaskFn references.
			for i := n; i < len(e.activeTasks); i++ {
				e.activeTasks[i] = nil
			}
			e.activeTasks = e.activeTasks[:n]
		}
	}
}

// drainIncoming moves all tasks currently residing in the incoming
// channel into the active execution slice.
func (e *Engine) drainIncoming() {
	for {
		select {
		case t := <-e.incoming:
			e.activeTasks = append(e.activeTasks, t)
		default:
			return
		}
	}
}

// runTaskSafely executes a TaskFn within a recovery block to ensure that
// individual task panics do not terminate the engine's worker goroutine.
// Panicked tasks are treated as finished and removed from the engine.
func (e *Engine) runTaskSafely(t TaskFn, dt time.Duration) (finished bool) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("Engine: TaskFn panicked", "error", r)
			finished = true
		}
	}()
	return t(dt)
}
