package util

import "sync"

type Event struct {
	Resource   string
	Action     string
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

func (e *EventRegistry) Register(resource, action string, handler HandlerFn) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if _, ok := e.registry[resource]; !ok {
		e.registry[resource] = make(map[string][]HandlerFn)
	}
	e.registry[resource][action] = append(e.registry[resource][action], handler)
}

func (e *EventRegistry) Emit(resource, action, identifier string) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	actions, ok := e.registry[resource]
	if !ok {
		return
	}

	for _, handler := range actions[action] {
		go handler(Event{Resource: resource, Action: action, Identifier: identifier})
	}
}
