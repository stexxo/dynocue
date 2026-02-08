package api

import (
	"log/slog"

	"gitlab.com/stexxo/dynocue/dynod/internal/bus"
)

type CueService struct {
	bus *bus.Client
}

func NewCueService(bus *bus.Client) *CueService {
	return &CueService{
		bus: bus,
	}
}

func (c *CueService) CreateCueList() bool {
	slog.Debug("creating new cue list")
	_, ok := c.bus.RequestHelper("show.cues.list.create", nil)
	return ok
}

func (c *CueService) GetCueLists() bool {
	slog.Debug("getting cue lists")
	return true
}
