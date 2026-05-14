package api

import (
	"errors"

	"github.com/stexxo/dynocue/components/cues/model"
	"github.com/stexxo/dynocue/core/messaging"
)

type CueingApi struct {
	model     *model.CueingModel
	messenger *messaging.Messenger
}

func NewCueingApi(model *model.CueingModel, messaging *messaging.Messenger) (*CueingApi, error) {
	c := &CueingApi{model: model, messenger: messaging}
	err := errors.Join(
		c.registerCueListApis(),
	)
	if err != nil {
		return nil, err
	}

	return c, nil
}
