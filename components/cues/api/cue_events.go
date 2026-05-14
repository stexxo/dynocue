package api

import (
	"github.com/stexxo/dynocue/components/cues/model"
	"github.com/stexxo/dynocue/util"
)

func (c *CueingApi) registerCueEvents() {
	c.model.RegisterEventHandler(model.ResourceCue, model.OperationCreated, eventHandler[CueChangeEvent](c.messenger, c.logger, c.CueCreated))
	c.model.RegisterEventHandler(model.ResourceCue, model.OperationUpdated, eventHandler[CueChangeEvent](c.messenger, c.logger, c.CueUpdated))
	c.model.RegisterEventHandler(model.ResourceCue, model.OperationDeleted, eventHandler[CueChangeEvent](c.messenger, c.logger, c.DeleteCueEvent))
}

const CueCreatedEventSubject = "event.cueing.cues.created"

type CueChangeEvent struct {
	CueId string `msgpack:"cueId" json:"cueId"`
}

func (c *CueingApi) CueCreated(ev util.Event) (string, *CueChangeEvent) {
	return CueCreatedEventSubject, &CueChangeEvent{CueId: ev.Identifier}
}

const CueAttributesUpdatedEventSubject = "event.cueing.cues.attributes.updated"

func (c *CueingApi) CueUpdated(ev util.Event) (string, *CueChangeEvent) {
	return CueAttributesUpdatedEventSubject, &CueChangeEvent{CueId: ev.Identifier}
}

const DeleteCueEventSubject = "event.cueing.cues.deleted"

func (c *CueingApi) DeleteCueEvent(ev util.Event) (string, *CueChangeEvent) {
	return DeleteCueEventSubject, &CueChangeEvent{CueId: ev.Identifier}
}
