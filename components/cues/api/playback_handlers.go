package api

import (
	"errors"

	"github.com/stexxo/dynocue/components/cues/model"
	"github.com/stexxo/dynocue/core/messaging"
)

func (c *CueingApi) registerPlaybackHandlers() error {
	return errors.Join(
		messaging.Reply[GoToCueRequest, GoToCueResponse](c.messenger, true, GoToCueRequestSubject, c.GoToCue),
		messaging.Reply[GoToNextCueRequest, GoToNextCueResponse](c.messenger, true, GoToNextCueRequestSubject, c.GoToNext),
	)
}

const GoToCueRequestSubject = "cueing.playback.go.to"

type GoToCueRequest struct {
	CueId string `msgpack:"cueId" json:"cueId"`
}

type GoToCueResponse struct{}

func (c *CueingApi) GoToCue(sub string, req *GoToCueRequest) (*GoToCueResponse, error) {
	err := c.engine.GoToCue(req.CueId)
	if errors.Is(err, model.ErrCueNotFound) {
		return nil, &messaging.FriendlyError{FriendlyErr: CueNotFound}
	}
	if err != nil {
		return nil, err
	}

	return nil, nil
}

const GoToNextCueRequestSubject = "cueing.playback.go.next"

type GoToNextCueRequest struct {
	CueListId string `msgpack:"cueListId" json:"cueListId"`
}

type GoToNextCueResponse struct{}

func (c *CueingApi) GoToNext(sub string, req *GoToNextCueRequest) (*GoToNextCueResponse, error) {
	err := c.engine.GoToNextCue(req.CueListId)
	if errors.Is(err, model.ErrCueNotFound) {
		return nil, &messaging.FriendlyError{FriendlyErr: CueNotFound}
	}
	if errors.Is(err, model.ErrCueListNotFound) {
		return nil, &messaging.FriendlyError{FriendlyErr: CueListNotFound}
	}
	if err != nil {
		return nil, err
	}
	return nil, nil
}
