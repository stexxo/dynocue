// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package engine

import (
	"context"
	"sync/atomic"
	"time"
)

type Task interface {
	Execute(t time.Duration) bool // Given the time since last execution, return true if task is done executing
}

type TaskEngine struct {
	interval time.Duration
	incoming chan Task
	cmds     chan string
	tasks    []Task

	ctx     context.Context
	cancel  context.CancelFunc
	running atomic.Bool
}

func NewEngine(ticksPerSecond int) *TaskEngine {
	msInterval := time.Second / time.Duration(ticksPerSecond) // Calculate interval in nanoseconds then save it as a duration

	return &TaskEngine{
		interval: msInterval,
		incoming: make(chan Task),
		tasks:    make([]Task, 100),
		cmds:     make(chan string),
		ctx:      nil,
	}
}

func (e *TaskEngine) Start() {
	if e.running.Load() {
		return
	}

	e.ctx, e.cancel = context.WithCancel(context.Background())
	go func() {
		defer e.running.Store(false)
		ticker := time.NewTicker(e.interval)
		lastExecuted := time.Now()
		for {
			select {
			case <-ticker.C:
				e.tick(time.Since(lastExecuted))
				lastExecuted = time.Now()
			case <-e.ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()

	e.running.Store(true)
}
func (e *TaskEngine) Stop() {
	if e.ctx == nil || !e.running.Load() {
		return
	}
	e.cancel()
}

func (e *TaskEngine) AddTask(task Task) {
	go func() {
		e.incoming <- task
	}()
}

func (e *TaskEngine) StopAll() {
	e.cmds <- "CLEAR_TASKS"
}

func (e *TaskEngine) tick(timeSinceLast time.Duration) {
	select {
	case <-e.ctx.Done():
		return
	case cmd := <-e.cmds:
		switch cmd {
		case "CLEAR_TASKS":
			e.tasks = e.tasks[:0]
		}
	}

	for {
		select {
		case <-e.ctx.Done():
			return
		case task := <-e.incoming:
			e.tasks = append(e.tasks, task)
		default:
			goto taskExecution
		}
	}

taskExecution:
	i := 0
	for j := 0; j < len(e.tasks); j++ {
		if finished := e.tasks[j].Execute(timeSinceLast); !finished {
			e.tasks[i] = e.tasks[j]
			i++
		}
	}
	clear(e.tasks[i:])
	e.tasks = e.tasks[:i]
}
