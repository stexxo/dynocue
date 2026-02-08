package manager

import (
	"errors"
	"log/slog"
	"sync/atomic"

	"gitlab.com/stexxo/dynocue/dynod/internal/bus"
	"gitlab.com/stexxo/dynocue/dynod/internal/show"
	"gitlab.com/stexxo/dynocue/dynod/internal/subsystems"
)

type Manager struct {
	evMgr *bus.Client
	show  *show.Show

	started atomic.Bool
}

func NewManager() *Manager {
	return &Manager{}
}

func (m *Manager) Start(client *bus.Client, show *show.Show) error {
	if m.started.Load() {
		return subsystems.ErrStarted
	}

	m.evMgr = client
	m.show = show

	err := errors.Join(
		m.evMgr.Subscribe("show.new", m.NewShow),
		m.evMgr.Subscribe("show.load", m.LoadShow),
		m.evMgr.Subscribe("show.close", m.CloseShow),
	)
	if err != nil {
		return err
	}

	m.started.Store(true)
	err = m.evMgr.Send(bus.Message{
		Subject: "subsystem.manager.status",
		Data:    []byte("STARTED"),
	})
	if err != nil {
		slog.Warn("failed to send subsystem manager status", "error", err)
	}
	return nil
}

func (m *Manager) Stop() error {
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

func (m *Manager) Name() string {
	return "manager"
}
