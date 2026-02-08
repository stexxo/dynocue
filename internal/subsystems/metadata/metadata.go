package metadata

import (
	"errors"
	"log/slog"
	"sync/atomic"

	"gitlab.com/stexxo/dynocue/dynod/internal/bus"
	"gitlab.com/stexxo/dynocue/dynod/internal/show"
	"gitlab.com/stexxo/dynocue/dynod/internal/subsystems"
)

const showMetadataKey = "metadata"

type Metadata struct {
	evMgr *bus.Client
	show  *show.Show

	started atomic.Bool
}

func NewMetadata() *Metadata {
	return &Metadata{}
}

func (m *Metadata) Start(client *bus.Client, show *show.Show) error {
	if m.started.Load() {
		return subsystems.ErrStarted
	}

	m.evMgr = client
	m.show = show

	err := errors.Join(
		m.evMgr.Subscribe("show.metadata.set.*", m.SetMetadataValue),
		m.evMgr.Subscribe("show.metadata.get.*", m.GetMetadataValue),
	)
	if err != nil {
		return err
	}

	m.started.Store(true)
	err = m.evMgr.Send(bus.Message{
		Subject: "subsystem.metadata.status",
		Data:    []byte("STARTED"),
	})
	if err != nil {
		slog.Warn("failed to send subsystem metadata status", "error", err)
	}
	return nil
}

func (m *Metadata) Stop() error {
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

func (m *Metadata) Name() string {
	return "metadata"
}
