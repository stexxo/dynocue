package api

import (
	"github.com/stexxo/dynocue/components/cues/model"
	"github.com/stexxo/dynocue/core/messaging"
)

type CueingApi struct {
	model     *model.CueingModel
	messenger *messaging.Messenger
}

func NewCueingApi(model *model.CueingModel, messaging *messaging.Messenger) *CueingApi {
	return &CueingApi{model: model, messenger: messaging}
}
