package api

import (
	"github.com/stexxo/dynocue/components/cues/model"
	"github.com/stexxo/dynocue/util"
)

func (c *CueingApi) registerActionEvents() {
	c.model.RegisterEventHandler(model.ResourceAction, model.OperationCreated, eventHandler[ActionChangeEvent](c.messenger, c.logger, c.ActionChanged))
	c.model.RegisterEventHandler(model.ResourceAction, model.OperationUpdated, eventHandler[ActionChangeEvent](c.messenger, c.logger, c.ActionChanged))
	c.model.RegisterEventHandler(model.ResourceAction, model.OperationDeleted, eventHandler[ActionChangeEvent](c.messenger, c.logger, c.ActionChanged))
}

const (
	ActionCreatedEventSubject           = "event.cueing.action.created"
	ActionAttributesUpdatedEventSubject = "event.cueing.action.attributes.updated"
	DeleteActionEventSubject            = "event.cueing.action.deleted"
)

type ActionChangeEvent struct {
	CueListId string `msgpack:"cueListId" json:"cueListId"`
	CueId     string `msgpack:"cueId" json:"cueId"`
	ActionId  string `msgpack:"actionId" json:"actionId"`
}

func (c *CueingApi) ActionChanged(ev util.Event) (string, *ActionChangeEvent) {
	var sub string
	switch ev.Operation {
	case model.OperationUpdated:
		sub = ActionAttributesUpdatedEventSubject
	case model.OperationDeleted:
		sub = DeleteActionEventSubject
	case model.OperationCreated:
		sub = ActionCreatedEventSubject
	}
	return sub, &ActionChangeEvent{
		CueListId: ev.EventData[model.MetadataCueListId],
		CueId:     ev.EventData[model.MetadataCueId],
		ActionId:  ev.EventData[model.MetadataActionId],
	}
}
