package cues

import (
	"encoding/json"
	"log/slog"

	"github.com/google/uuid"
	"gitlab.com/stexxo/dynocue/dynod/internal/bus"
	"gitlab.com/stexxo/dynocue/dynod/internal/cues"
)

func (c *CueSystem) CreateCueList(msg bus.Message) {
	slog.Debug("received request for cue lists")
	var response []byte
	defer func() {
		if msg.Reply != "" {
			c.evMgr.RespondHelper(msg, response)
		}
	}()

	db, err := c.show.GetDatabase()
	if err != nil {
		slog.Error("failed to get database", "error", err)
		return
	}

	id := uuid.NewString()
	number := float64(1)
	cl, err := cues.GetCueLists(db)
	if len(cl) > 0 {
		number = cl[len(cl)-1].Number + 1
	}

	err = cues.CreateCueList(db, number, id)
	if err != nil {
		response = []byte("FAILED")
		return
	}

	c.evMgr.SendHelper("show.cues.lists.updated", []byte(id))
}

func (c *CueSystem) GetCueLists(msg bus.Message) {
	slog.Debug("received request for cue lists")
	var response []byte
	defer func() {
		if msg.Reply != "" {
			c.evMgr.RespondHelper(msg, response)
		}
	}()

	db, err := c.show.GetDatabase()
	if err != nil {
		slog.Error("failed to get database", "error", err)
		response = []byte("FAILED")
		return
	}

	cueLists, err := cues.GetCueLists(db)
	if err != nil {
		slog.Error("failed to get cue lists", "error", err)
		response = []byte("FAILED")
		return
	}

	b, err := json.Marshal(cueLists)
	if err != nil {
		slog.Error("failed to marshal cue lists", "error", err)
		response = []byte("FAILED")
		return
	}

	response = b
}
