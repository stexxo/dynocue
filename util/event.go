package util

import (
	"slices"
	"sync"
)

type Event struct {
	Resource  string
	Operation string
	EventData map[string]string
}

type HandlerFn func(Event)

type EventRegistry struct {
	mu       sync.RWMutex
	registry map[string]map[string][]HandlerFn
}

func NewEventRegistry() *EventRegistry {
	return &EventRegistry{
		registry: make(map[string]map[string][]HandlerFn),
	}
}

func (e *EventRegistry) Register(resource, operation string, handler HandlerFn) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if _, ok := e.registry[resource]; !ok {
		e.registry[resource] = make(map[string][]HandlerFn)
	}
	e.registry[resource][operation] = append(e.registry[resource][operation], handler)
}

func (e *EventRegistry) Emit(resource, operation string, metadata ...string) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	operations, ok := e.registry[resource]
	if !ok {
		return
	}

	ev := Event{Resource: resource, Operation: operation}
	if len(metadata) > 0 {
		ev.EventData = make(map[string]string)
		for pair := range slices.Chunk(metadata, 2) {
			ev.EventData[pair[0]] = pair[1]
		}
	}
	for _, handler := range operations[operation] {
		go handler(ev)
	}
}
