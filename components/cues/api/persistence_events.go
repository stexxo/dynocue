package api

import (
	"github.com/stexxo/dynocue/components/cues/model"
	"github.com/stexxo/dynocue/util"
)

func (c *CueingApi) registerPersistenceEvents() {
	c.model.RegisterEventHandler(model.ResourceModel, model.OperationLoaded, eventHandler[PersistenceChangeEvent](c.messenger, c.logger, c.PersistenceChanged))
}

const (
	ModelLoadedEventSubject = "event.cueing.persistence.loaded"
)

type PersistenceChangeEvent struct {
}

func (c *CueingApi) PersistenceChanged(ev util.Event) (string, *PersistenceChangeEvent) {
	var sub string
	switch ev.Operation {
	case model.OperationLoaded:
		sub = ModelLoadedEventSubject
	}
	return sub, &PersistenceChangeEvent{}
}
