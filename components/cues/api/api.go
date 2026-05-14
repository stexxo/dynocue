package api

import (
	"errors"

	"github.com/stexxo/dynocue/components/cues/model"
	"github.com/stexxo/dynocue/core/logging"
	"github.com/stexxo/dynocue/core/messaging"
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
	)
	if err != nil {
		return nil, err
	}

	c.registerCueListEvents()

	return c, nil
}
