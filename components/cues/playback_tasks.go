// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cues

import (
	"math"
	"time"

	"github.com/stexxo/dynocue/components/cues/types"
	"github.com/stexxo/dynocue/core/logging"
	"github.com/stexxo/dynocue/core/messaging"
	"github.com/stexxo/dynocue/engine"
)

const (
	ExecutionStageInitialized = iota
	ExecutionStageDelayed
	ExecutionStageActions
	ExecutionStageFollow
	ExecutionStageComplete
)

const CueExecutionStatusEventSubject = "event.cueing.execution.status"

type CueExecutionStatusEvent struct {
	CueListId string                 `msgpack:"cueListId" json:"cueListId"`
	CueId     string                 `msgpack:"cueId" json:"cueId"`
	Stage     int                    `msgpack:"stage" json:"stage"`
	Data      map[string]interface{} `msgpack:"data" json:"data"`
}

type ExecuteCueTask struct {
	cue *types.Cue

	messenger *messaging.Messenger
	logger    logging.Logger
	engine    *engine.TaskEngine

	executionStage int

	delayElapsed  time.Duration
	actionElapsed time.Duration
	followElapsed time.Duration

	actionCursor float64
}

func NewExecuteCueTask(cue *types.Cue, messenger *messaging.Messenger, logger logging.Logger, e *engine.TaskEngine) *ExecuteCueTask {
	return &ExecuteCueTask{
		cue:       cue,
		messenger: messenger,
		logger:    logger,
		engine:    e,
	}
}

func (e *ExecuteCueTask) Execute(t time.Duration) bool {
	var next bool
	switch e.executionStage {
	case ExecutionStageInitialized:
		e.executionStage = ExecutionStageDelayed
	case ExecutionStageDelayed:
		next = e.delayStage(t)
	case ExecutionStageActions:
		next = e.actionStage(t)
	case ExecutionStageFollow:
		next = e.followStage(t)
	case ExecutionStageComplete:
		return true
	}

	if next {
		e.executionStage++
	}

	return false
}

func (e *ExecuteCueTask) delayStage(t time.Duration) bool {
	err := messaging.Publish(e.messenger, CueExecutionStatusEventSubject, &CueExecutionStatusEvent{
		CueListId: e.cue.Metadata.CueListId,
		CueId:     e.cue.Metadata.CueId,
		Stage:     e.executionStage,
		Data: map[string]interface{}{
			"delay":   e.cue.Metadata.Delay,
			"elapsed": int(math.Min(float64(e.delayElapsed), float64(e.cue.Metadata.Delay))),
		},
	})
	if err != nil {
		e.logger.Error("failed to emit status event about cue delay")
	}

	if e.delayElapsed > e.cue.Metadata.Delay {
		return true
	}

	return false
}

func (e *ExecuteCueTask) actionStage(duration time.Duration) bool {

	return false
}

func (e *ExecuteCueTask) followStage(duration time.Duration) bool {
	return false
}

const ActionExecutionStatusEventSubject = "event.cueing.execution.status"

type ActionExecutionStatusEvent struct {
	CueListId string                 `msgpack:"cueListId" json:"cueListId"`
	CueId     string                 `msgpack:"cueId" json:"cueId"`
	ActionId  string                 `msgpack:"actionId" json:"actionId"`
	Stage     int                    `msgpack:"stage" json:"stage"`
	Data      map[string]interface{} `msgpack:"data" json:"data"`
}
