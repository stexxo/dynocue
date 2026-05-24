package api

import (
	"errors"

	"github.com/stexxo/dynocue/components/audio/model"
	"github.com/stexxo/dynocue/components/system"
	"github.com/stexxo/dynocue/core/logging"
	"github.com/stexxo/dynocue/core/messaging"
	"github.com/stexxo/dynocue/util"
)

type AudioAPI struct {
	model       *model.AudioModel
	persistence *system.PersistenceManager
	messenger   *messaging.Messenger
	logger      logging.Logger
}

func NewAudioAPI(m *model.AudioModel, persistence *system.PersistenceManager, messenger *messaging.Messenger, logger logging.Logger) (*AudioAPI, error) {
	a := &AudioAPI{
		model:       m,
		persistence: persistence,
		messenger:   messenger,
		logger:      logger,
	}

	err := errors.Join(
		a.registerFileApis(),
	)
	if err != nil {
		return nil, err
	}

	a.registerFileEvents()

	return a, nil
}

func eventHandler[T any](m *messaging.Messenger, l logging.Logger, evFn func(util.Event) (string, *T)) util.HandlerFn {
	return func(event util.Event) {
		sub, body := evFn(event)
		err := messaging.Publish(m, sub, body)
		if err != nil {
			l.Error("failed to publish event", "error", err, "resource", event.Resource, "operation", event.Operation)
		}
	}
}
