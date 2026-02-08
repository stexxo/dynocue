package api

import (
	"gitlab.com/stexxo/dynocue/dynod/internal/bus"
)

type LifecycleService struct {
	bus *bus.Client
}

func NewLifecycleService(bus *bus.Client) *LifecycleService {
	return &LifecycleService{
		bus: bus,
	}
}

func (l *LifecycleService) NewShow() bool {
	if resp, ok := l.bus.RequestHelper("gui.show.new", nil); !ok || string(resp.Data) != "SUCCESS" {
		return false
	}
	return true
}

func (l *LifecycleService) OpenShow() bool {
	if resp, ok := l.bus.RequestHelper("gui.show.load", nil); !ok || string(resp.Data) != "SUCCESS" {
		return false
	}
	return true
}

func (l *LifecycleService) CloseShow(windowName string) bool {
	if resp, ok := l.bus.RequestHelper("gui.show.close", []byte(windowName)); !ok || string(resp.Data) != "SUCCESS" {
		return false
	}
	return true
}
