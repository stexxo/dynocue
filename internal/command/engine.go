package engine

import (
	"context"
	"log/slog"
	"sync"
)

// Command defines the functional signature for executable tasks.
type Command func()

// CommandEngine facilitates serialized execution of tasks on a dedicated background goroutine.
// This ensures that state mutations are performed sequentially, maintaining manager integrity.
type CommandEngine struct {
	commands chan Command

	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
	mu      sync.Mutex
	running bool
}

// NewCommandEngine initializes a CommandEngine with a specified queue capacity.
func NewCommandEngine(bufferSize int) *CommandEngine {
	return &CommandEngine{
		commands: make(chan Command, bufferSize),
	}
}

// Start initializes the background worker goroutine. This method is idempotent;
// if the engine is already active, the call returns immediately.
func (e *CommandEngine) Start() {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.running {
		return
	}

	e.ctx, e.cancel = context.WithCancel(context.Background())
	e.running = true
	e.wg.Add(1)

	go e.processLoop()
	slog.Info("CommandEngine: Worker goroutine started")
}

// Stop signals the worker goroutine to terminate. This method is idempotent.
// Following the C++ logic, the loop exits as soon as the signal is received.
func (e *CommandEngine) Stop() {
	e.mu.Lock()
	if !e.running {
		e.mu.Unlock()
		return
	}
	e.running = false
	cancel := e.cancel
	e.mu.Unlock()

	if cancel != nil {
		cancel()
	}

	e.wg.Wait()
	slog.Info("CommandEngine: Worker goroutine joined")
}

// AddCommand appends a task to the execution queue. If the queue is full,
// the command is dropped to prevent blocking the caller.
func (e *CommandEngine) AddCommand(cmd Command) {
	if cmd == nil {
		slog.Warn("CommandEngine: Attempted to add a nil command")
		return
	}

	e.mu.Lock()
	running := e.running
	e.mu.Unlock()

	if !running {
		return
	}

	select {
	case e.commands <- cmd:
		// Command successfully queued
	default:
		slog.Warn("CommandEngine: Queue capacity reached; dropping command")
	}
}

// processLoop provides the execution context for the serialized command queue.
func (e *CommandEngine) processLoop() {
	defer e.wg.Done()

	for {
		select {
		case <-e.ctx.Done():
			return
		case cmd := <-e.commands:
			e.runSafely(cmd)
		}
	}
}

// runSafely executes the command within a recovery block to protect
// the worker goroutine from subsystem panics.
func (e *CommandEngine) runSafely(cmd Command) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("CommandEngine: Panic caught during command execution", "error", r)
		}
	}()
	cmd()
}
