package util

import "sync"

type Event struct {
	Resource   string
	Operation  string
	Identifier string
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

func (e *EventRegistry) Emit(resource, operation, identifier string) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	operations, ok := e.registry[resource]
	if !ok {
		return
	}

	for _, handler := range operations[operation] {
		go handler(Event{Resource: resource, Operation: operation, Identifier: identifier})
	}
}
