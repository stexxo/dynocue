// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package api

import (
	"errors"

	"github.com/stexxo/dynocue/components/cues/model"
	"github.com/stexxo/dynocue/core/messaging"
	"github.com/stexxo/dynocue/db"
)

func (c *CueingApi) registerExecutionApis() error {
	return errors.Join(
		messaging.Reply[GoToCueRequest, GoToCueResponse](c.messenger, true, GoToCueRequestSubject, c.GoToCue),
		messaging.Reply[GoToNextCueRequest, GoToNextCueResponse](c.messenger, true, GoToNextCueRequestSubject, c.GoToNextCue),
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
