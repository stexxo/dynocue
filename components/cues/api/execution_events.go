// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package api

import (
	"github.com/stexxo/dynocue/components/cues/model"
	"github.com/stexxo/dynocue/util"
)

func (c *CueingApi) registerExecutionEvents() {
	c.model.RegisterEventHandler(model.ResourceCueExecution, model.OperationStarted, eventHandler[ExecutionChangeEvent](c.messenger, c.logger, c.ExecutionChanged))
	c.model.RegisterEventHandler(model.ResourceCueExecution, model.OperationFinished, eventHandler[ExecutionChangeEvent](c.messenger, c.logger, c.ExecutionChanged))
	c.model.RegisterEventHandler(model.ResourceCueExecution, model.OperationUnselected, eventHandler[ExecutionChangeEvent](c.messenger, c.logger, c.ExecutionChanged))
	c.model.RegisterEventHandler(model.ResourceCueExecution, model.OperationDeleted, eventHandler[ExecutionChangeEvent](c.messenger, c.logger, c.ExecutionChanged))
	c.model.RegisterEventHandler(model.ResourceCueExecution, model.OperationUpdated, eventHandler[ExecutionChangeEvent](c.messenger, c.logger, c.ExecutionChanged))
}

const (
	ExecutionStartedEventSubject    = "event.cueing.execution.started"
	ExecutionFinishedEventSubject   = "event.cueing.execution.finished"
	ExecutionUnselectedEventSubject = "event.cueing.execution.unselected"
	ExecutionDeletedEventSubject    = "event.cueing.execution.deleted"
	ExecutionUpdatedEventSubject    = "event.cueing.execution.updated"
)

type ExecutionChangeEvent struct {
	CueListId string `msgpack:"cueListId" json:"cueListId"`
	CueId     string `msgpack:"cueId" json:"cueId"`
}

func (c *CueingApi) ExecutionChanged(ev util.Event) (string, *ExecutionChangeEvent) {
	var sub string
	switch ev.Operation {
	case model.OperationStarted:
		sub = ExecutionStartedEventSubject
	case model.OperationFinished:
		sub = ExecutionFinishedEventSubject
	case model.OperationUnselected:
		sub = ExecutionUnselectedEventSubject
	case model.OperationDeleted:
		sub = ExecutionDeletedEventSubject
	case model.OperationUpdated:
		sub = ExecutionUpdatedEventSubject
	}
	return sub, &ExecutionChangeEvent{
		CueListId: ev.EventData[model.MetadataCueListId],
		CueId:     ev.EventData[model.MetadataCueId],
	}
}
