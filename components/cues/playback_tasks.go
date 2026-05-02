// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cues

import (
	"slices"
	"sync"
	"time"

	"github.com/stexxo/dynocue/components/cues/types"
	"github.com/stexxo/dynocue/core/logging"
	"github.com/stexxo/dynocue/core/messaging"
	"github.com/stexxo/dynocue/engine"
)

const (
	CueExecutionStageInitialized = iota
	CueExecutionStageDelayed
	CueExecutionStageActionStart
	CueExecutionStageActionWait
	CueExecutionStageFollow
	CueExecutionStageComplete
)

const CueExecutionStatusEventSubject = "event.cueing.execution.status"

type CueExecutionStatusEvent struct {
	CueListId      string        `msgpack:"cueListId" json:"cueListId"`
	CueId          string        `msgpack:"cueId" json:"cueId"`
	Stage          int           `msgpack:"stage" json:"stage"`
	TotalElapsed   time.Duration `msgpack:"totalElapsed" json:"totalElapsed"`
	ElapsedInStage time.Duration `msgpack:"elapsedinstage" json:"elapsedinstage"`
}

type ExecuteCueTask struct {
	cueList *types.CueList
	cue     *types.Cue

	messenger *messaging.Messenger
	logger    logging.Logger
	engine    *engine.TaskEngine

	executionStage     int
	timeElapsedInStage time.Duration
	totalElapsed       time.Duration

	actionsStarted bool
	actionTasks    []*ExecuteActionTask
}

func NewExecuteCueTask(cue *types.Cue, messenger *messaging.Messenger, logger logging.Logger, e *engine.TaskEngine) *ExecuteCueTask {
	return &ExecuteCueTask{
		cue:            cue,
		messenger:      messenger,
		logger:         logger,
		engine:         e,
		executionStage: CueExecutionStageInitialized,
	}
}

func (e *ExecuteCueTask) Execute(t time.Duration) bool {
	e.timeElapsedInStage += t
	e.totalElapsed += t

	if e.executionStage == CueExecutionStageInitialized {
		e.executionStage = CueExecutionStageDelayed
	}

	if e.executionStage == CueExecutionStageDelayed {
		if e.timeElapsedInStage > e.cue.Attributes.Delay {
			e.executionStage = CueExecutionStageActionStart
		}
	}

	if e.executionStage == CueExecutionStageActionStart {
		for _, action := range e.cue.Actions {
			actionTask := NewExecuteActionTask(&action, e.messenger, e.logger)
			e.engine.AddTask(actionTask)
		}
		e.executionStage = CueExecutionStageActionWait
	}

	if e.executionStage == CueExecutionStageActionWait {
		e.actionTasks = slices.DeleteFunc(e.actionTasks, func(action *ExecuteActionTask) bool {
			action.mu.Lock()
			defer action.mu.Unlock()

			return action.executionState == ActionExecutionStageComplete
		})

		if len(e.actionTasks) == 0 {
			e.executionStage = CueExecutionStageFollow
		}
	}

	if e.executionStage == CueExecutionStageFollow {
		if e.cue.Attributes.Follow == 0 {
			e.executionStage = CueExecutionStageComplete
		}

		if e.timeElapsedInStage > e.cue.Attributes.Follow {
			nextCue := e.cueList.Cues.GetNextByNumber(e.cue.Num())
			if nextCue != nil {
				task := NewExecuteCueTask(*nextCue, e.messenger, e.logger, e.engine)
				e.engine.AddTask(task)
			}
			e.executionStage = CueExecutionStageComplete
		}
	}

	err := messaging.Publish(e.messenger, CueExecutionStatusEventSubject, CueExecutionStatusEvent{
		CueListId:      e.cue.Attributes.CueListId,
		CueId:          e.cue.Attributes.CueId,
		Stage:          e.executionStage,
		ElapsedInStage: e.timeElapsedInStage,
		TotalElapsed:   e.totalElapsed,
	})
	if err != nil {
		e.logger.Error("failed to publish cue execution status event", "err", err)
	}

	if e.executionStage == ActionExecutionStageComplete {
		return true
	}

	return false
}

const ActionExecutionStatusEventSubject = "event.cueing.execution.status"

const (
	ActionExecutionStageInitialized = iota
	ActionExecutionStageDelayed
	ActionExecutionStageRunning
	ActionExecutionStageComplete
)

type ActionExecutionStatusEvent struct {
	CueListId string                 `msgpack:"cueListId" json:"cueListId"`
	CueId     string                 `msgpack:"cueId" json:"cueId"`
	ActionId  string                 `msgpack:"actionId" json:"actionId"`
	Stage     int                    `msgpack:"stage" json:"stage"`
	Data      map[string]interface{} `msgpack:"data" json:"data"`
}

type ExecuteActionTask struct {
	mu sync.RWMutex

	action    *types.Action
	messenger *messaging.Messenger
	logger    logging.Logger

	executionState     int
	timeElapsedInStage time.Duration
	totalElapsed       time.Duration
}

func NewExecuteActionTask(action *types.Action, messenger *messaging.Messenger, logger logging.Logger) *ExecuteActionTask {
	return &ExecuteActionTask{
		action:         action,
		messenger:      messenger,
		logger:         logger,
		executionState: ActionExecutionStageInitialized,
	}
}

func (e *ExecuteActionTask) Execute(t time.Duration) bool {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.timeElapsedInStage += t
	e.totalElapsed += t

	if e.executionState == ActionExecutionStageInitialized {
		e.executionState = CueExecutionStageDelayed
	}

	if e.executionState == ActionExecutionStageDelayed {
		if e.timeElapsedInStage > e.action.Delay {
			e.executionState = ActionExecutionStageRunning
		}
	}

	if e.executionState == ActionExecutionStageRunning {
		body := map[string]interface{}{}
		for _, f := range e.action.Fields {
			body[f.FieldName] = f.Value
		}
		err := messaging.Publish(e.messenger, e.action.Subject, body)
		if err != nil {
			e.logger.Error("failed to publish action execution status event", "err", err)
		}
		e.executionState = CueExecutionStageComplete
	}

	err := messaging.Publish(e.messenger, ActionExecutionStatusEventSubject, ActionExecutionStatusEvent{
		CueListId: e.action.CueListId,
		CueId:     e.action.CueId,
		Stage:     e.executionState,
	})

	if err != nil {
		e.logger.Error("failed to publish action execution status event", "err", err)
	}

	if e.executionState == ActionExecutionStageComplete {
		return true
	}

	return false
}
