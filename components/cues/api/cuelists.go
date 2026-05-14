package api

import (
	"errors"

	"github.com/stexxo/dynocue/components/cues/model"
	"github.com/stexxo/dynocue/core/messaging"
)

const (
	CueListNumberExists = "Cue list number already exists"
	CueListNotFound     = "Cue list not found"
)

func (c *CueingApi) registerCueListApis() error {
	return errors.Join(
		messaging.Reply[CreateCueListRequest, CreateCueListResponse](c.messenger, true, CreateCueListRequestSubject, c.CreateCueList),
	)
}

const CreateCueListRequestSubject = "request.cueing.cuelists.create"

type CreateCueListRequest struct {
	Number      uint   `msgpack:"number" json:"number" validate:"gte=0"`
	CueListType string `msgpack:"cueListType" json:"cueListType" validate:"required,oneof=SEQUENTIAL"`
}

type CreateCueListResponse struct {
	Id     string `msgpack:"id" json:"id"`
	Number uint   `msgpack:"number" json:"number"`
}

func (c *CueingApi) CreateCueList(sub string, request *CreateCueListRequest) (*CreateCueListResponse, error) {
	id, num, err := c.model.CreateCueList(request.Number, request.CueListType)
	if errors.Is(err, model.ErrNumberExists) {
		return nil, errors.Join(err, &messaging.FriendlyError{FriendlyErr: CueListNumberExists})
	}
	if err != nil {
		return nil, err
	}
	return &CreateCueListResponse{Id: id, Number: num}, nil
}
