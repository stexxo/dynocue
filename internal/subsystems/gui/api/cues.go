package api

import (
	"encoding/json"
	"log/slog"

	"gitlab.com/stexxo/dynocue/dynod/internal/bus"
	"gitlab.com/stexxo/dynocue/dynod/internal/cues"
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

func (c *CueService) GetCueLists() ([]cues.CueListNumber, bool) {
	slog.Debug("getting cue lists")
	resp, ok := c.bus.RequestHelper("show.cues.lists.getall", nil)
	if !ok {
		return nil, false
	}

	out := make([]cues.CueListNumber, 0)
	err := json.Unmarshal(resp.Data, &out)
	if err != nil {
		return nil, false
	}

	return out, true
}
