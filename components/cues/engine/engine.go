package engine

import (
	"github.com/stexxo/dynocue/components/cues/model"
	"github.com/stexxo/dynocue/core/logging"
)

type CueingEngine struct {
	model  *model.CueingModel
	logger logging.Logger
}

func NewCueingEngine(m *model.CueingModel) *CueingEngine {
	return &CueingEngine{
		model: m,
	}
}
