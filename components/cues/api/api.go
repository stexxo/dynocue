package api

import (
	"errors"

	"github.com/stexxo/dynocue/components/cues/model"
	"github.com/stexxo/dynocue/core/logging"
	"github.com/stexxo/dynocue/core/messaging"
	"github.com/stexxo/dynocue/util"
)

type CueingApi struct {
	model     *model.CueingModel
	messenger *messaging.Messenger
	logger    logging.Logger
}

func NewCueingApi(model *model.CueingModel, messaging *messaging.Messenger, logger logging.Logger) (*CueingApi, error) {
	c := &CueingApi{model: model, messenger: messaging, logger: logger}
	err := errors.Join(
		c.registerCueListApis(),
		c.registerCueApis(),
		c.registerActionApis(),
	)
	if err != nil {
		return nil, err
	}

	c.registerCueListEvents()
	c.registerCueEvents()
	c.registerActionEvents()

	return c, nil
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
