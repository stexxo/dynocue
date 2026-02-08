package cues

import (
	"log/slog"

	"gitlab.com/stexxo/dynocue/dynod/internal/bus"
)

func (c *CueSystem) CreateCueList(msg bus.Message) {
	slog.Debug("received request for cue lists")
	var response []byte
	defer func() {
		if msg.Reply != "" {
			c.evMgr.RespondHelper(msg, response)
		}
	}()
}

func (c *CueSystem) GetCueLists(msg bus.Message) {
	slog.Debug("received request for cue lists")
	var response []byte
	defer func() {
		if msg.Reply != "" {
			c.evMgr.RespondHelper(msg, response)
		}
	}()
}
