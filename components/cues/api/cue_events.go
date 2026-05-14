package api

import (
	"github.com/stexxo/dynocue/components/cues/model"
	"github.com/stexxo/dynocue/util"
)

func (c *CueingApi) registerCueEvents() {
	c.model.RegisterEventHandler(model.ResourceCue, model.OperationCreated, eventHandler[CueChangeEvent](c.messenger, c.logger, c.CueChanged))
	c.model.RegisterEventHandler(model.ResourceCue, model.OperationUpdated, eventHandler[CueChangeEvent](c.messenger, c.logger, c.CueChanged))
	c.model.RegisterEventHandler(model.ResourceCue, model.OperationDeleted, eventHandler[CueChangeEvent](c.messenger, c.logger, c.CueChanged))
}

const (
	CueCreatedEventSubject           = "event.cueing.cues.created"
	CueAttributesUpdatedEventSubject = "event.cueing.cues.attributes.updated"
	DeleteCueEventSubject            = "event.cueing.cues.deleted"
)

type CueChangeEvent struct {
	CueListId string `msgpack:"cueListId" json:"cueListId"`
	CueId     string `msgpack:"cueId" json:"cueId"`
}

func (c *CueingApi) CueChanged(ev util.Event) (string, *CueChangeEvent) {
	var sub string
	switch ev.Operation {
	case model.OperationUpdated:
		sub = CueAttributesUpdatedEventSubject
	case model.OperationDeleted:
		sub = DeleteCueEventSubject
	case model.OperationCreated:
		sub = CueCreatedEventSubject
	}
	return sub, &CueChangeEvent{CueListId: ev.EventData[model.MetadataCueListId], CueId: ev.EventData[model.MetadataCueId]}
}
