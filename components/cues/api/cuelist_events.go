package api

import (
	"github.com/stexxo/dynocue/components/cues/model"
	"github.com/stexxo/dynocue/util"
)

func (c *CueingApi) registerCueListEvents() {
	c.model.RegisterEventHandler(model.ResourceCueList, model.OperationCreated, eventHandler[CueListChangeEvent](c.messenger, c.logger, c.CueListChanged))
	c.model.RegisterEventHandler(model.ResourceCueList, model.OperationUpdated, eventHandler[CueListChangeEvent](c.messenger, c.logger, c.CueListChanged))
	c.model.RegisterEventHandler(model.ResourceCueList, model.OperationDeleted, eventHandler[CueListChangeEvent](c.messenger, c.logger, c.CueListChanged))
}

const (
	CueListCreatedEventSubject           = "event.cueing.cuelists.created"
	CueListAttributesUpdatedEventSubject = "event.cueing.cuelists.updated"
	DeleteCueListEventSubject            = "event.cueing.cuelists.deleted"
)

type CueListChangeEvent struct {
	CueListId string `msgpack:"cueListId" json:"cueListId"`
}

func (c *CueingApi) CueListChanged(ev util.Event) (string, *CueListChangeEvent) {
	var sub string
	switch ev.Operation {
	case model.OperationUpdated:
		sub = CueListAttributesUpdatedEventSubject
	case model.OperationDeleted:
		sub = DeleteCueListEventSubject
	case model.OperationCreated:
		sub = CueListCreatedEventSubject
	}
	return sub, &CueListChangeEvent{CueListId: ev.EventData[model.MetadataCueListId]}
}
