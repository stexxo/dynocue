// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package api

import (
	"errors"

	"github.com/stexxo/dynocue/components/cues/model"
	"github.com/stexxo/dynocue/components/cues/types"
	"github.com/stexxo/dynocue/core/messaging"
)

const (
	CueNotFound     = "Cue not found"
	CueNumberExists = "Cue number already exists"
)

func (c *CueingApi) registerCueApis() error {
	return errors.Join(
		messaging.Reply[CreateCueRequest, CreateCueResponse](c.messenger, true, CreateCueRequestSubject, c.CreateCue),
		messaging.Reply[EnumerateCuesRequest, EnumerateCuesResponse](c.messenger, true, EnumerateCuesRequestSubject, c.EnumerateCues),
		messaging.Reply[GetCueByNumberRequest, GetCueByNumberResponse](c.messenger, true, GetCueByNumberRequestSubject, c.GetCueByNumber),
		messaging.Reply[GetCueByIdRequest, GetCueByIdResponse](c.messenger, true, GetCueByIdRequestSubject, c.GetCueById),
		messaging.Reply[DeleteCueRequest, DeleteCueResponse](c.messenger, true, DeleteCueRequestSubject, c.DeleteCue),
		messaging.Reply[UpdateCueAttributesRequest, UpdateCueAttributesResponse](c.messenger, true, UpdateCueAttributesRequestSubject, c.UpdateCueAttributes),
	)
}

const CreateCueRequestSubject = "request.cueing.cue.create"

type CreateCueRequest struct {
	CueListId string `msgpack:"cueListId" json:"cueListId" validate:"required"`
	Number    uint   `msgpack:"number" json:"number" validate:"gte=0"`
}

type CreateCueResponse struct {
	CueId  string `msgpack:"cueId" json:"cueId"`
	Number uint   `msgpack:"number" json:"number"`
}

func (c *CueingApi) CreateCue(sub string, request *CreateCueRequest) (*CreateCueResponse, error) {
	id, num, err := c.model.CreateCue(request.CueListId, request.Number)
	if errors.Is(err, model.ErrNumberExists) {
		return nil, errors.Join(err, &messaging.FriendlyError{FriendlyErr: CueNumberExists})
	}
	if err != nil {
		return nil, err
	}
	return &CreateCueResponse{CueId: id, Number: num}, nil
}

const EnumerateCuesRequestSubject = "request.cueing.cue.enumerate"

type EnumerateCuesRequest struct {
	CueListId string `msgpack:"cueListId" json:"cueListId" validate:"required"`
}

type EnumerateCuesResponse struct {
	Cues []types.Cue `msgpack:"cues" json:"cues"`
}

func (c *CueingApi) EnumerateCues(sub string, request *EnumerateCuesRequest) (*EnumerateCuesResponse, error) {
	cues, err := c.model.EnumerateCues(request.CueListId)
	if err != nil {
		return nil, err
	}
	return &EnumerateCuesResponse{Cues: cues}, nil
}

const GetCueByNumberRequestSubject = "request.cueing.cue.get.number"

type GetCueByNumberRequest struct {
	CueListId string  `msgpack:"cueListId" json:"cueListId" validate:"required"`
	Number    float64 `msgpack:"number" json:"number" validate:"required,gt=0"`
}

type GetCueByNumberResponse struct {
	Cue types.Cue `msgpack:"cue" json:"cue"`
}

func (c *CueingApi) GetCueByNumber(sub string, request *GetCueByNumberRequest) (*GetCueByNumberResponse, error) {
	out, err := c.model.GetCueByNumber(request.CueListId, uint(request.Number))
	if errors.Is(err, model.ErrCueNotFound) {
		return nil, errors.Join(err, &messaging.FriendlyError{FriendlyErr: CueNotFound})
	}
	if err != nil {
		return nil, err
	}
	return &GetCueByNumberResponse{
		Cue: *out,
	}, nil
}

const GetCueByIdRequestSubject = "request.cueing.cue.get.id"

type GetCueByIdRequest struct {
	CueId string `msgpack:"cueId" json:"cueId" validate:"required"`
}

type GetCueByIdResponse struct {
	Cue types.Cue `msgpack:"cue" json:"cue"`
}

func (c *CueingApi) GetCueById(sub string, request *GetCueByIdRequest) (*GetCueByIdResponse, error) {
	out, err := c.model.GetCueById(request.CueId)
	if errors.Is(err, model.ErrCueNotFound) {
		return nil, errors.Join(err, &messaging.FriendlyError{FriendlyErr: CueNotFound})
	}
	if err != nil {
		return nil, err
	}

	return &GetCueByIdResponse{
		Cue: *out,
	}, nil
}

const DeleteCueRequestSubject = "request.cueing.cue.delete"

type DeleteCueRequest struct {
	CueId string `msgpack:"cueId" json:"cueId" validate:"required"`
}

type DeleteCueResponse struct{}

func (c *CueingApi) DeleteCue(sub string, request *DeleteCueRequest) (*DeleteCueResponse, error) {
	err := c.model.DeleteCueById(request.CueId)
	if err != nil {
		return nil, err
	}

	return &DeleteCueResponse{}, nil
}

const UpdateCueAttributesRequestSubject = "request.cueing.cue.attributes.update"

type UpdateCueAttributesRequest struct {
	CueId string      `msgpack:"cueId" json:"cueId" validate:"required"`
	Field string      `msgpack:"field" json:"field" validate:"required,ne=cueId,ne=number,ne=cueListId"`
	Value interface{} `msgpack:"value" json:"value" validate:"required"`
}

type UpdateCueAttributesResponse struct{}

func (c *CueingApi) UpdateCueAttributes(sub string, request *UpdateCueAttributesRequest) (*UpdateCueAttributesResponse, error) {
	err := c.model.UpdateCueAttribute(request.CueId, request.Field, request.Value)
	if errors.Is(err, model.ErrCueNotFound) {
		return nil, errors.Join(err, &messaging.FriendlyError{FriendlyErr: CueNotFound})
	}
	if err != nil {
		return nil, err
	}

	return &UpdateCueAttributesResponse{}, nil
}
