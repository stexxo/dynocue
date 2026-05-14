package api

import (
	"github.com/stexxo/dynocue/components/cues/model"
	"github.com/stexxo/dynocue/util"
)

func (c *CueingApi) registerCueListEvents() {
	c.model.RegisterEventHandler(model.ResourceCueList, model.OperationCreated, eventHandler[CueListChangeEvent](c.messenger, c.logger, c.CueListCreated))
	c.model.RegisterEventHandler(model.ResourceCueList, model.OperationUpdated, eventHandler[CueListChangeEvent](c.messenger, c.logger, c.CueListUpdated))
	c.model.RegisterEventHandler(model.ResourceCueList, model.OperationDeleted, eventHandler[CueListChangeEvent](c.messenger, c.logger, c.DeleteCueListEvent))
}

const CueListCreatedEventSubject = "event.cueing.cuelists.created"

type CueListChangeEvent struct {
	CueListId string `msgpack:"cueListId" json:"cueListId"`
}

func (c *CueingApi) CueListCreated(ev util.Event) (string, *CueListChangeEvent) {
	return CueListCreatedEventSubject, &CueListChangeEvent{CueListId: ev.Identifier}
}

const CueListAttributesUpdatedEventSubject = "event.cueing.cuelists.updated"

func (c *CueingApi) CueListUpdated(ev util.Event) (string, *CueListChangeEvent) {
	return CueListAttributesUpdatedEventSubject, &CueListChangeEvent{CueListId: ev.Identifier}
}

const DeleteCueListEventSubject = "event.cueing.cuelists.deleted"

func (c *CueingApi) DeleteCueListEvent(ev util.Event) (string, *CueListChangeEvent) {
	return DeleteCueListEventSubject, &CueListChangeEvent{CueListId: ev.Identifier}
}
