package api

import (
	"github.com/stexxo/dynocue/components/cues/model"
	"github.com/stexxo/dynocue/core/logging"
	"github.com/stexxo/dynocue/core/messaging"
	"github.com/stexxo/dynocue/util"
)

func (c *CueingApi) registerCueListEvents() {
	c.model.RegisterEventHandler(model.ResourceCueList, model.OperationCreated, handler[CueListChangeEvent](c.messenger, c.logger, c.CueListCreated))
	c.model.RegisterEventHandler(model.ResourceCueList, model.OperationUpdated, handler[CueListChangeEvent](c.messenger, c.logger, c.CueListUpdated))
	c.model.RegisterEventHandler(model.ResourceCueList, model.OperationDeleted, handler[CueListChangeEvent](c.messenger, c.logger, c.DeleteCueListEvent))
}

func handler[T any](m *messaging.Messenger, l logging.Logger, evFn func(util.Event) (string, *T)) util.HandlerFn {
	return func(event util.Event) {
		sub, body := evFn(event)
		err := messaging.Publish(m, sub, body)
		if err != nil {
			l.Error("failed to publish event", "error", err, "resource", event.Resource, "operation", event.Operation, "identifier", event.Identifier)
		}
	}
}

const CueListCreatedEventSubject = "event.cueing.cuelists.created"

type CueListChangeEvent struct {
	CueListId string `msgpack:"cueListId" json:"cueListId"`
}

func (c *CueingApi) CueListCreated(ev util.Event) (string, *CueListChangeEvent) {
	return CueListCreatedEventSubject, &CueListChangeEvent{CueListId: ev.Identifier}
}

const CueListAttributesUpdatedEventSubject = "event.cueing.cuelists.attributes.updated"

func (c *CueingApi) CueListUpdated(ev util.Event) (string, *CueListChangeEvent) {
	return CueListAttributesUpdatedEventSubject, &CueListChangeEvent{CueListId: ev.Identifier}
}

const DeleteCueListEventSubject = "event.cueing.cuelists.deleted"

func (c *CueingApi) DeleteCueListEvent(ev util.Event) (string, *CueListChangeEvent) {
	return DeleteCueListEventSubject, &CueListChangeEvent{CueListId: ev.Identifier}
}
