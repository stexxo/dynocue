package api

import (
	"github.com/stexxo/dynocue/components/audio/model"
	"github.com/stexxo/dynocue/components/system"
	"github.com/stexxo/dynocue/core/logging"
	"github.com/stexxo/dynocue/core/messaging"
)

type AudioAPI struct {
	model       *model.AudioModel
	persistence *system.PersistenceManager
	messenger   *messaging.Messenger
	logger      *logging.Logger
}

func NewAudioAPI(m *model.AudioModel, persistence *system.PersistenceManager, messenger *messaging.Messenger, logger *logging.Logger) *AudioAPI {
	return &AudioAPI{
		model:       m,
		persistence: persistence,
		messenger:   messenger,
		logger:      logger,
	}
}
