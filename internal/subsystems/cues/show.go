package cues

import (
	"errors"
	"log/slog"
	"sync/atomic"

	"gitlab.com/stexxo/dynocue/dynod/internal/bus"
	"gitlab.com/stexxo/dynocue/dynod/internal/show"
	"gitlab.com/stexxo/dynocue/dynod/internal/subsystems"
)

type CueSystem struct {
	evMgr *bus.Client
	show  *show.Show

	started atomic.Bool
}

func NewCueSystem() *CueSystem {
	return &CueSystem{}
}

func (m *CueSystem) Start(client *bus.Client, show *show.Show) error {
	if m.started.Load() {
		return subsystems.ErrStarted
	}

	m.evMgr = client
	m.show = show

	err := errors.Join(
		m.evMgr.Subscribe("show.cues.list.create", m.CreateCueList),
		m.evMgr.Subscribe("show.cues.lists.getall", m.GetCueLists),
	)
	if err != nil {
		return err
	}

	m.started.Store(true)
	err = m.evMgr.Send(bus.Message{
		Subject: "subsystem.cues.status",
		Data:    []byte("STARTED"),
	})
	if err != nil {
		slog.Warn("failed to send subsystem manager status", "error", err)
	}
	return nil
}

func (m *CueSystem) Stop() error {
	if !m.started.Load() {
		return subsystems.ErrStopped
	}
	err := errors.Join(m.evMgr.Stop())
	if err == nil {
		m.evMgr = nil
		m.show = nil
	}
	m.started.Store(false)

	return err
}

func (m *CueSystem) Name() string {
	return "cues"
}
