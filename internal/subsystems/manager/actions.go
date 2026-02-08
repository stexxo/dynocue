package manager

import (
	"log/slog"
	"strconv"

	"gitlab.com/stexxo/dynocue/dynod/internal/bus"
)

func (m *Manager) NewShow(msg bus.Message) {
	var response string
	defer func() {
		if msg.Reply != "" {
			m.evMgr.RespondHelper(msg, []byte(response))
		}
	}()
	slog.Debug("received request for new show")

	err := m.show.NewShow(string(msg.Data))
	if err != nil {
		response = "FAILED"
	} else {
		response = "SUCCESS"
		slog.Info("New Show Initialized Successfully")
	}
}

func (m *Manager) LoadShow(msg bus.Message) {
	var response string
	defer func() {
		if msg.Reply != "" {
			m.evMgr.RespondHelper(msg, []byte(response))
		}
	}()
	slog.Debug("received request for loading show")

	err := m.show.Load(string(msg.Data))
	if err != nil {
		response = "FAILED"
	} else {
		response = "SUCCESS"
		slog.Info("Loaded Show Initialized Successfully")
	}
}

func (m *Manager) CloseShow(msg bus.Message) {
	var response string
	defer func() {
		if msg.Reply != "" {
			m.evMgr.RespondHelper(msg, []byte(response))
		}
	}()
	m.show.Close()
	slog.Debug("received request for closing show")
}

func (m *Manager) IsShowLoaded(msg bus.Message) {
	m.evMgr.RespondHelper(msg, []byte(strconv.FormatBool(m.show.IsInitialized())))
}
