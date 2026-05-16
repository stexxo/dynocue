// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package api

import (
	"github.com/stexxo/dynocue/components/cues/model"
	"github.com/stexxo/dynocue/util"
)

func (c *CueingApi) registerActionExecutionEvents() {
	c.model.RegisterEventHandler(model.ResourceActionExecution, model.OperationStarted, eventHandler[ActionExecutionChangeEvent](c.messenger, c.logger, c.ActionExecutionChanged))
	c.model.RegisterEventHandler(model.ResourceActionExecution, model.OperationDeleted, eventHandler[ActionExecutionChangeEvent](c.messenger, c.logger, c.ActionExecutionChanged))
	c.model.RegisterEventHandler(model.ResourceActionExecution, model.OperationUpdated, eventHandler[ActionExecutionChangeEvent](c.messenger, c.logger, c.ActionExecutionChanged))
}

const (
	ActionExecutionStartedEventSubject = "event.cueing.execution.action.started"
	ActionExecutionDeletedEventSubject = "event.cueing.execution.action.deleted"
	ActionExecutionUpdatedEventSubject = "event.cueing.execution.action.updated"
)

type ActionExecutionChangeEvent struct {
	CueListId string `msgpack:"cueListId" json:"cueListId"`
	CueId     string `msgpack:"cueId" json:"cueId"`
	ActionId  string `msgpack:"actionId" json:"actionId"`
}

func (c *CueingApi) ActionExecutionChanged(ev util.Event) (string, *ActionExecutionChangeEvent) {
	var sub string
	switch ev.Operation {
	case model.OperationStarted:
		sub = ActionExecutionStartedEventSubject
	case model.OperationDeleted:
		sub = ActionExecutionDeletedEventSubject
	case model.OperationUpdated:
		sub = ActionExecutionUpdatedEventSubject
	}
	return sub, &ActionExecutionChangeEvent{
		CueListId: ev.EventData[model.MetadataCueListId],
		CueId:     ev.EventData[model.MetadataCueId],
		ActionId:  ev.EventData[model.MetadataActionId],
	}
}
