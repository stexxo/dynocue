// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cues

import (
	"errors"

	"github.com/stexxo/dynocue/components/cues/types"
	"github.com/stexxo/dynocue/core/messaging"
	"github.com/stexxo/dynocue/util"
)

const CueListNumberExists = "Cue List Number Already Exists"
const CueListNotFound = "Cue List Not Found."

// CreateCueList

const CreateCueListRequestSubject = "request.cueing.cuelists.create"
const CueListCreatedEventSubject = "event.cueing.cuelists.created"

type CreateCueListRequest struct {
	Number      float64 `msgpack:"number" json:"number" validate:"gte=0"`
	CueListType string  `msgpack:"cueListType" json:"cueListType" validate:"required,oneof=SEQUENTIAL"`
}

type CreateCueListResponse struct {
	Id     string  `msgpack:"id" json:"id"`
	Number float64 `msgpack:"number" json:"number"`
}

type CueListCreatedEvent struct {
	CueListMetadata types.CueListMetadata `msgpack:"cueListMetadata" json:"cueListMetadata"`
}

func (p *Cueing) CreateCueList(sub string, request *CreateCueListRequest) (*CreateCueListResponse, error) {
	c := types.NewCueList(request.Number, request.CueListType)
	ok := p.model.CueLists.Add(c)
	if !ok {
		return nil, &messaging.FriendlyError{FriendlyErr: CueListNumberExists}
	}

	err := messaging.Publish(p.Messenger(), CueListCreatedEventSubject, &CueListCreatedEvent{
		CueListMetadata: c.Metadata,
	})
	if err != nil {
		p.Logger().Error("Failed to publish cue list created event", "error", err, "cueListNumber", request.Number, "cueListType", request.CueListType)
		return nil, err
	}

	return &CreateCueListResponse{
		Id:     c.Id(),
		Number: c.Num(),
	}, nil
}

// EnumerateCueLists

const EnumerateCueListsRequestSubject = "request.cueing.cuelists.enumerate"

type EnumerateCueListsRequest struct{}

type EnumerateCueListsResponse struct {
	CueLists []types.CueListMetadata `msgpack:"cueLists" json:"cueLists"`
}

type CueListEnumeration struct {
	Number      float64 `msgpack:"number" json:"number"`
	Label       string  `msgpack:"label" json:"label"`
	CueListType string  `msgpack:"cueListType" json:"cueListType"`
}

func (p *Cueing) EnumerateCueLists(sub string, request *EnumerateCueListsRequest) (*EnumerateCueListsResponse, error) {
	out := make([]types.CueListMetadata, 0, p.model.CueLists.Len())
	p.model.CueLists.ForEach(func(list *types.CueList) {
		out = append(out, list.Metadata)
	})

	return &EnumerateCueListsResponse{CueLists: out}, nil
}

// GetCueListByNumber

const GetCueListByNumberRequestSubject = "request.cueing.cuelists.get.number"

type GetCueListByNumberRequest struct {
	Number float64 `msgpack:"number" json:"number" validate:"required,gt=0"`
}

type GetCueListByNumberResponse struct {
	CueListMetadata types.CueListMetadata `msgpack:"cueListMetadata" json:"cueListMetadata"`
}

func (p *Cueing) GetCueListByNumber(sub string, request *GetCueListByNumberRequest) (*GetCueListByNumberResponse, error) {
	out := p.model.CueLists.GetFunc(func(list *types.CueList) bool {
		return list.Num() == request.Number
	})
	if out == nil {
		return nil, &messaging.FriendlyError{FriendlyErr: CueListNotFound}
	}

	return &GetCueListByNumberResponse{
		CueListMetadata: (*out).Metadata,
	}, nil
}

// GetCueListById

const GetCueListByIdRequestSubject = "request.cueing.cuelists.get.id"

type GetCueListByIdRequest struct {
	Id string `msgpack:"id" json:"id" validate:"required"`
}

type GetCueListByIdResponse struct {
	CueListMetadata types.CueListMetadata `msgpack:"cueListMetadata" json:"cueListMetadata"`
}

func (p *Cueing) GetCueListById(sub string, request *GetCueListByIdRequest) (*GetCueListByIdResponse, error) {
	cl, err := p.getCueListById(request.Id)
	if err != nil {
		return nil, err
	}

	return &GetCueListByIdResponse{
		CueListMetadata: cl.Metadata,
	}, nil
}

func (p *Cueing) getCueListById(id string) (*types.CueList, error) {
	cl := p.model.CueLists.GetFunc(func(list *types.CueList) bool {
		return list.Id() == id
	})
	if cl == nil {
		return nil, &messaging.FriendlyError{FriendlyErr: CueListNotFound}
	}

	return *cl, nil
}

// RenumberCueList

const RenumberCueListRequestSubject = "request.cueing.cuelists.renumber"
const RenumberCueListEventSubject = "event.cueing.cuelists.renumber"

type RenumberCueListsRequest struct {
	Id        string  `msgpack:"id" json:"id" validate:"required"`
	NewNumber float64 `msgpack:"newNumber" json:"newNumber" validate:"required,gt=0"`
}

type RenumberCueListsResponse struct{}

type RenumberCueListEvent struct {
	Id        string  `msgpack:"id" json:"id" validate:"required"`
	NewNumber float64 `msgpack:"newNumber" json:"newNumber"`
}

func (p *Cueing) RenumberCueList(sub string, request *RenumberCueListsRequest) (*RenumberCueListsResponse, error) {
	err := p.model.CueLists.MoveFunc(func(list *types.CueList) bool {
		return list.Id() == request.Id
	}, request.NewNumber)
	if errors.Is(err, util.ErrNotFound) {
		p.Logger().Debug("could not renumber cue list because the original cue list does not exist", "err", err, "cueListId", request.Id, "newNumber", request.NewNumber)
		return nil, &messaging.FriendlyError{FriendlyErr: CueListNotFound}
	}

	if errors.Is(err, util.ErrExists) {
		p.Logger().Debug("could not renumber cue list because the new number already exists", "err", err, "cueListId", request.Id, "newNumber", request.NewNumber)
		return nil, &messaging.FriendlyError{FriendlyErr: CueListNumberExists}
	}

	if err != nil {
		p.Logger().Error("Failed to renumber cue list", "err", err, "cueListId", request.Id, "newNumber", request.NewNumber)
		return nil, err
	}

	err = messaging.Publish(p.Messenger(), RenumberCueListEventSubject, &RenumberCueListEvent{
		Id:        request.Id,
		NewNumber: request.NewNumber,
	})
	if err != nil {
		p.Logger().Error("Failed to publish renumber cue list event", "err", err, "cuelistId", request.Id, "newNumber", request.NewNumber)
		return nil, err
	}

	return &RenumberCueListsResponse{}, nil
}

// DeleteCueList

const DeleteCueListRequestSubject = "request.cueing.cuelists.delete"
const DeleteCueListEventSubject = "event.cueing.cuelists.deleted"

type DeleteCueListsRequest struct {
	Id string `msgpack:"id" json:"id" validate:"required"`
}

type DeleteCueListsResponse struct{}

type CueListDeletedEvent struct {
	Id string `msgpack:"id" json:"id"`
}

func (p *Cueing) DeleteCueList(sub string, request *DeleteCueListsRequest) (*DeleteCueListsResponse, error) {
	p.model.CueLists.RemoveFunc(func(list *types.CueList) bool {
		return list.Id() == request.Id
	})

	err := messaging.Publish(p.Messenger(), DeleteCueListEventSubject, &CueListDeletedEvent{
		Id: request.Id,
	})
	if err != nil {
		p.Logger().Error("failed to publish cue list deleted event", "error", err, "id", request.Id)
		return nil, err
	}

	return &DeleteCueListsResponse{}, nil
}

// Update Operations

// Update Events

const CueListMetadataUpdatedEventSubject = "event.cueing.cuelists.metadata.updated"

type CueListMetadataUpdatedEvent struct {
	Metadata types.CueListMetadata `msgpack:"metadata" json:"metadata"`
}

// UpdateCueListMetadata

const UpdateCueListMetadataRequestSubject = "request.cueing.cuelists.metadata.update"

type UpdateCueListMetadataRequest struct {
	Id    string      `msgpack:"id" json:"id" validate:"required"`
	Field string      `msgpack:"field" json:"field" validate:"required,ne=id,ne=number,ne=cueListType"`
	Value interface{} `msgpack:"value" json:"value" validate:"required"`
}

type UpdateCueListMetadataResponse struct {
	Metadata types.CueListMetadata `msgpack:"metadata" json:"metadata"`
}

func (p *Cueing) UpdateCueListMetadata(sub string, request *UpdateCueListMetadataRequest) (*UpdateCueListMetadataResponse, error) {
	cl, err := p.getCueListById(request.Id)
	if err != nil {
		return nil, err
	}

	err = util.UpdateStructByTag("json", request.Field, request.Value, &cl.Metadata)
	if err != nil {
		p.Logger().Error("failed to update field in cuelist", "err", err, "field", request.Field, "cueListId", request.Id)
		return nil, err
	}

	err = messaging.Publish(p.Messenger(), CueListMetadataUpdatedEventSubject, &CueListMetadataUpdatedEvent{
		Metadata: cl.Metadata,
	})
	if err != nil {
		p.Logger().Error("Failed to publish updated cue list metadata", "error", err)
		return nil, err
	}

	return &UpdateCueListMetadataResponse{Metadata: cl.Metadata}, nil
}
