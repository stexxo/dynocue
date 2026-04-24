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

const CueNotFound = "Cue Not Found"
const CueNumberExists = "Cue Number Already Exists"

// CreateCue

const CreateCueRequestSubject = "request.cueing.cue.create"
const CueCreatedEventSubject = "event.cueing.cue.created"

type CreateCueRequest struct {
	CueListId string  `msgpack:"cueListId" json:"cueListId"`
	CueNumber float64 `msgpack:"cueNumber" json:"cueNumber"`
}

type CreateCueResponse struct {
	CueListId string  `msgpack:"cueListId" json:"cueListId"`
	CueId     string  `msgpack:"cueId" json:"cueId"`
	CueNumber float64 `msgpack:"cueNumber" json:"cueNumber"`
}

type CueCreatedEvent struct {
	CueListId string            `msgpack:"cueListId" json:"cueListId"`
	Metadata  types.CueMetadata `msgpack:"metadata" json:"metadata"`
}

func (p *Cueing) CreateCue(sub string, req *CreateCueRequest) (*CreateCueResponse, error) {
	cue := types.NewCue(req.CueListId, req.CueNumber)

	cl, err := p.getCueListById(req.CueListId)
	if err != nil {
		return nil, err
	}

	if cl == nil {
		return nil, &messaging.FriendlyError{FriendlyErr: CueListNotFound}
	}

	cl.Cues.Add(cue)

	err = messaging.Publish(p.Messenger(), CueCreatedEventSubject, &CueCreatedEvent{
		CueListId: req.CueListId,
		Metadata:  cue.Metadata,
	})
	if err != nil {
		p.Logger().Error("failed to publish cue created event", "error", err)
		return nil, err
	}

	return &CreateCueResponse{CueListId: req.CueListId, CueId: cue.Metadata.CueId, CueNumber: cue.Metadata.Number}, nil
}

// EnumerateCues

const EnumerateCuesRequestSubject = "request.cueing.cue.enumerate"

type EnumerateCuesRequest struct {
	CueListId string `msgpack:"cueListId" json:"cueListId" validate:"required"`
}

type EnumerateCuesResponse struct {
	Cues []types.CueMetadata `msgpack:"cues" json:"cues"`
}

func (p *Cueing) EnumerateCues(sub string, request *EnumerateCuesRequest) (*EnumerateCuesResponse, error) {
	cl, err := p.getCueListById(request.CueListId)
	if err != nil {
		return nil, err
	}

	out := make([]types.CueMetadata, 0, cl.Cues.Len())
	cl.Cues.ForEach(func(cue *types.Cue) {
		out = append(out, cue.Metadata)
	})

	return &EnumerateCuesResponse{Cues: out}, nil
}

// GetCueByNumber

const GetCueByNumberRequestSubject = "request.cueing.cue.get.number"

type GetCueByNumberRequest struct {
	CueListNumber float64 `msgpack:"cueListNumber" json:"cueListNumber" validate:"required,gt=0"`
	CueNumber     float64 `msgpack:"cueNumber" json:"cueNumber" validate:"required,gt=0"`
}

type GetCueByNumberResponse struct {
	Metadata types.CueMetadata `msgpack:"metadata" json:"metadata"`
}

func (p *Cueing) GetCueByNumber(sub string, request *GetCueByNumberRequest) (*GetCueByNumberResponse, error) {
	clPtr := p.model.CueLists.GetFunc(func(list *types.CueList) bool {
		return list.Num() == request.CueListNumber
	})
	if clPtr == nil {
		return nil, &messaging.FriendlyError{FriendlyErr: CueListNotFound}
	}
	cl := *clPtr

	out := cl.Cues.GetFunc(func(cue *types.Cue) bool {
		return cue.Num() == request.CueNumber
	})
	if out == nil {
		return nil, &messaging.FriendlyError{FriendlyErr: CueNotFound}
	}

	cue := *out

	return &GetCueByNumberResponse{Metadata: cue.Metadata}, nil
}

// GetCueById

const GetCueByIdRequestSubject = "request.cueing.cue.get.id"

type GetCueByIdRequest struct {
	CueListId string `msgpack:"cueListId" json:"cueListId" validate:"required"`
	CueId     string `msgpack:"cueId" json:"cueId" validate:"required"`
}

type GetCueByIdResponse struct {
	Metadata types.CueMetadata `msgpack:"metadata" json:"metadata"`
}

func (p *Cueing) GetCueById(sub string, request *GetCueByIdRequest) (*GetCueByIdResponse, error) {
	cl, err := p.getCueListById(request.CueListId)
	if err != nil {
		return nil, err
	}

	cue, err := p.getCueById(cl, request.CueId)
	if err != nil {
		return nil, err
	}

	return &GetCueByIdResponse{Metadata: cue.Metadata}, nil
}

func (p *Cueing) getCueById(cl *types.CueList, id string) (*types.Cue, error) {
	cue := cl.Cues.GetFunc(func(c *types.Cue) bool {
		return c.Metadata.CueId == id
	})
	if cue == nil {
		return nil, &messaging.FriendlyError{FriendlyErr: CueNotFound}
	}

	return *cue, nil
}

// RenumberCue

const RenumberCueRequestSubject = "request.cueing.cue.renumber"
const RenumberCueEventSubject = "event.cueing.cue.renumber"

type RenumberCueRequest struct {
	CueListId string  `msgpack:"cueListId" json:"cueListId" validate:"required"`
	CueId     string  `msgpack:"cueId" json:"cueId" validate:"required"`
	NewNumber float64 `msgpack:"newNumber" json:"newNumber" validate:"required,gt=0"`
}

type RenumberCueResponse struct{}

type RenumberCueEvent struct {
	CueListId string  `msgpack:"cueListId" json:"cueListId"`
	CueId     string  `msgpack:"cueId" json:"cueId"`
	NewNumber float64 `msgpack:"newNumber" json:"newNumber"`
}

func (p *Cueing) RenumberCue(sub string, request *RenumberCueRequest) (*RenumberCueResponse, error) {
	cl, err := p.getCueListById(request.CueListId)
	if err != nil {
		return nil, err
	}

	err = cl.Cues.MoveFunc(func(cue *types.Cue) bool {
		return cue.Metadata.CueId == request.CueId
	}, request.NewNumber)
	if errors.Is(err, util.ErrNotFound) {
		return nil, &messaging.FriendlyError{FriendlyErr: CueNotFound}
	}

	if errors.Is(err, util.ErrExists) {
		return nil, &messaging.FriendlyError{FriendlyErr: CueNumberExists}
	}

	if err != nil {
		p.Logger().Error("Failed to renumber cue", "err", err, "cueListId", request.CueListId, "cueId", request.CueId, "newNumber", request.NewNumber)
		return nil, err
	}

	err = messaging.Publish(p.Messenger(), RenumberCueEventSubject, &RenumberCueEvent{
		CueListId: request.CueListId,
		CueId:     request.CueId,
		NewNumber: request.NewNumber,
	})
	if err != nil {
		p.Logger().Error("Failed to publish renumber cue event", "err", err, "cueListId", request.CueListId, "cueId", request.CueId, "newNumber", request.NewNumber)
		return nil, err
	}

	return &RenumberCueResponse{}, nil
}

// DeleteCue

const DeleteCueRequestSubject = "request.cueing.cue.delete"
const DeleteCueEventSubject = "event.cueing.cue.deleted"

type DeleteCueRequest struct {
	CueListId string `msgpack:"cueListId" json:"cueListId" validate:"required"`
	CueId     string `msgpack:"cueId" json:"cueId" validate:"required"`
}

type DeleteCueResponse struct{}

type CueDeletedEvent struct {
	CueListId string `msgpack:"cueListId" json:"cueListId"`
	CueId     string `msgpack:"cueId" json:"cueId"`
}

func (p *Cueing) DeleteCue(sub string, request *DeleteCueRequest) (*DeleteCueResponse, error) {
	cl, err := p.getCueListById(request.CueListId)
	if err != nil {
		return nil, err
	}

	cl.Cues.RemoveFunc(func(cue *types.Cue) bool {
		return cue.Metadata.CueId == request.CueId
	})

	err = messaging.Publish(p.Messenger(), DeleteCueEventSubject, &CueDeletedEvent{
		CueListId: request.CueListId,
		CueId:     request.CueId,
	})
	if err != nil {
		p.Logger().Error("failed to publish cue deleted event", "error", err, "cueListId", request.CueListId, "id", request.CueId)
		return nil, err
	}

	return &DeleteCueResponse{}, nil
}

// Update Operations

// Update Events

const CueMetadataUpdatedEventSubject = "event.cueing.cue.metadata.updated"

type CueMetadataUpdatedEvent struct {
	Metadata types.CueMetadata `msgpack:"metadata" json:"metadata"`
}

// UpdateCueMetadata

const UpdateCueMetadataRequestSubject = "request.cueing.cue.metadata.update"

type UpdateCueMetadataRequest struct {
	CueListId string      `msgpack:"cueListId" json:"cueListId" validate:"required"`
	CueId     string      `msgpack:"cueId" json:"cueId" validate:"required,ne=cueListId,ne=cueId,ne=number"`
	Field     string      `msgpack:"field" json:"field" validate:"required"`
	Value     interface{} `msgpack:"value" json:"value"`
}

type UpdateCueMetadataResponse struct {
	Metadata types.CueMetadata `msgpack:"metadata" json:"metadata"`
}

func (p *Cueing) UpdateCueMetadata(sub string, request *UpdateCueMetadataRequest) (*UpdateCueMetadataResponse, error) {
	cl, err := p.getCueListById(request.CueListId)
	if err != nil {
		return nil, err
	}

	cue, err := p.getCueById(cl, request.CueId)
	if err != nil {
		return nil, err
	}

	err = util.UpdateStructByTag("json", request.Field, request.Value, &cue.Metadata)
	if err != nil {
		p.Logger().Error("failed to update field in cue")
		return nil, err
	}

	err = messaging.Publish(p.Messenger(), CueMetadataUpdatedEventSubject, &CueMetadataUpdatedEvent{
		Metadata: cue.Metadata,
	})
	if err != nil {
		p.Logger().Error("Failed to publish updated cue label", "error", err)
		return nil, err
	}

	return &UpdateCueMetadataResponse{Metadata: cue.Metadata}, nil
}
