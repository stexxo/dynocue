// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

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
	ActionCreatedEventSubject           = "event.cueing.actions.created"
	ActionAttributesUpdatedEventSubject = "event.cueing.actions.updated"
	ActionDeletedEventSubject           = "event.cueing.actions.deleted"
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
		sub = ActionDeletedEventSubject
	case model.OperationCreated:
		sub = ActionCreatedEventSubject
	}
	return sub, &ActionChangeEvent{
		CueListId: ev.EventData[model.MetadataCueListId],
		CueId:     ev.EventData[model.MetadataCueId],
		ActionId:  ev.EventData[model.MetadataActionId],
	}
}
