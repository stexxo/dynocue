// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package api

import (
	"errors"

	"github.com/stexxo/dynocue/components/cues/model"
	"github.com/stexxo/dynocue/components/cues/types"
	"github.com/stexxo/dynocue/core/messaging"
	"github.com/stexxo/dynocue/db"
)

func (c *CueingApi) registerExecutionApis() error {
	return errors.Join(
		messaging.Reply[GoToCueRequest, GoToCueResponse](c.messenger, true, GoToCueRequestSubject, c.GoToCue),
		messaging.Reply[GoToNextCueRequest, GoToNextCueResponse](c.messenger, true, GoToNextCueRequestSubject, c.GoToNextCue),
		messaging.Reply[GetSelectedCueRequest, GetSelectedCueResponse](c.messenger, true, GetSelectedCueRequestSubject, c.GetSelectedCue),
		messaging.Reply[GetCueExecutionRequest, GetCueExecutionResponse](c.messenger, true, GetCueExecutionRequestSubject, c.GetCueExecution),
		messaging.Reply[EnumerateCueExecutionsRequest, EnumerateCueExecutionsResponse](c.messenger, true, EnumerateCueExecutionsRequestSubject, c.EnumerateCueExecutions),
	)
}

const (
	NoNextCue      = "No next cue found"
	NoCueSelected  = "No cue selected"
)

const GoToCueRequestSubject = "request.cueing.execution.goto.cue"

type GoToCueRequest struct {
	CueId string `msgpack:"cueId" json:"cueId" validate:"required"`
}

type GoToCueResponse struct{}

func (c *CueingApi) GoToCue(sub string, request *GoToCueRequest) (*GoToCueResponse, error) {
	err := c.engine.GoToCue(request.CueId)
	if errors.Is(err, model.ErrCueNotFound) {
		return nil, errors.Join(err, &messaging.FriendlyError{FriendlyErr: CueNotFound})
	}
	if errors.Is(err, model.ErrCueListNotFound) {
		return nil, errors.Join(err, &messaging.FriendlyError{FriendlyErr: CueListNotFound})
	}
	if err != nil {
		return nil, err
	}
	return &GoToCueResponse{}, nil
}

const GoToNextCueRequestSubject = "request.cueing.execution.goto.next"

type GoToNextCueRequest struct {
	CueListId string `msgpack:"cueListId" json:"cueListId" validate:"required"`
}

type GoToNextCueResponse struct{}

func (c *CueingApi) GoToNextCue(sub string, request *GoToNextCueRequest) (*GoToNextCueResponse, error) {
	err := c.engine.GoToNextCue(request.CueListId)
	if errors.Is(err, model.ErrCueListNotFound) {
		return nil, errors.Join(err, &messaging.FriendlyError{FriendlyErr: CueListNotFound})
	}
	if errors.Is(err, model.ErrNoNextCue) {
		return nil, errors.Join(err, &messaging.FriendlyError{FriendlyErr: NoNextCue})
	}
	if errors.Is(err, db.ErrItemNotFound) {
		return nil, errors.Join(err, &messaging.FriendlyError{FriendlyErr: NoCueSelected})
	}
	if err != nil {
		return nil, err
	}
	return &GoToNextCueResponse{}, nil
}

const GetSelectedCueRequestSubject = "request.cueing.execution.get.selected"

type GetSelectedCueRequest struct {
	CueListId string `msgpack:"cueListId" json:"cueListId" validate:"required"`
}

type GetSelectedCueResponse struct {
	Execution *types.CueExecution `msgpack:"execution" json:"execution"`
}

func (c *CueingApi) GetSelectedCue(sub string, request *GetSelectedCueRequest) (*GetSelectedCueResponse, error) {
	res, err := c.model.GetSelectedCue(request.CueListId)
	if errors.Is(err, db.ErrItemNotFound) {
		return &GetSelectedCueResponse{Execution: nil}, nil
	}
	if err != nil {
		return nil, err
	}
	return &GetSelectedCueResponse{Execution: res}, nil
}

const GetCueExecutionRequestSubject = "request.cueing.execution.get.id"

type GetCueExecutionRequest struct {
	CueId string `msgpack:"cueId" json:"cueId" validate:"required"`
}

type GetCueExecutionResponse struct {
	Execution *types.CueExecution `msgpack:"execution" json:"execution"`
}

func (c *CueingApi) GetCueExecution(sub string, request *GetCueExecutionRequest) (*GetCueExecutionResponse, error) {
	res, err := c.model.GetCueExecution(request.CueId)
	if errors.Is(err, model.ErrCueNotFound) {
		return nil, errors.Join(err, &messaging.FriendlyError{FriendlyErr: CueNotFound})
	}
	if err != nil {
		return nil, err
	}
	return &GetCueExecutionResponse{Execution: res}, nil
}

const EnumerateCueExecutionsRequestSubject = "request.cueing.execution.enumerate"

type EnumerateCueExecutionsRequest struct {
	CueListId string `msgpack:"cueListId" json:"cueListId" validate:"required"`
}

type EnumerateCueExecutionsResponse struct {
	Executions []types.CueExecution `msgpack:"executions" json:"executions"`
}

func (c *CueingApi) EnumerateCueExecutions(sub string, request *EnumerateCueExecutionsRequest) (*EnumerateCueExecutionsResponse, error) {
	res, err := c.model.EnumerateCueExecutions(request.CueListId)
	if err != nil {
		return nil, err
	}
	return &EnumerateCueExecutionsResponse{Executions: res}, nil
}
