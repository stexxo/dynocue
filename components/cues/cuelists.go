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

const CreateCueListRequestSubject = "request.cueing.cuelists.create"
const CueListCreatedEventSubject = "event.cueing.cuelists.created"

type CreateCueListRequest struct {
	Number      float64 `msgpack:"number" json:"number" validate:"gte=0"`
	CueListType string  `msgpack:"cueListType" json:"cueListType" validate:"required,oneof=SEQUENTIAL"`
}

type CreateCueListResponse struct {
	Number float64 `msgpack:"number" json:"number"`
}

type CueListCreatedEvent struct {
	CueListMetadata types.CueListMetadata `msgpack:"cueListMetadata" json:"cueListMetadata"`
}

const CueListNumberExists = "Cue List Number Already Exists"

func (p *Cueing) CreateCueList(sub string, request *CreateCueListRequest) (*CreateCueListResponse, error) {
	c := &types.CueList{Metadata: types.CueListMetadata{Number: request.Number, CueListType: request.CueListType}}
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
		Number: c.Num(),
	}, nil
}

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

const GetCueListRequestSubject = "request.cueing.cuelists.get"

type GetCueListRequest struct {
	Number float64 `msgpack:"number" json:"number" validate:"required,gt=0"`
}

type GetCueListResponse struct {
	CueListMetadata types.CueListMetadata `msgpack:"cueListMetadata" json:"cueListMetadata"`
}

const CueListNotFound = "Cue List Not Found."

func (p *Cueing) GetCueList(sub string, request *GetCueListRequest) (*GetCueListResponse, error) {
	out := p.model.CueLists.Get(request.Number)
	if out == nil {
		return nil, &messaging.FriendlyError{FriendlyErr: CueListNotFound}
	}

	return &GetCueListResponse{
		CueListMetadata: (*out).Metadata,
	}, nil
}

const CueListMetadataUpdatedEventSubject = "event.cueing.cuelists.metadata.updated"

type CueListMetadataUpdatedEvent struct {
	Metadata types.CueListMetadata `msgpack:"metadata" json:"metadata"`
}

const UpdateCueListLabelRequestSubject = "request.cueing.cuelists.updateLabel"

type UpdateCueListLabelRequest struct {
	Number float64 `msgpack:"number" json:"number" validate:"required,gt=0"`
	Label  string  `msgpack:"label" json:"label"`
}
type UpdateCueListLabelResponse struct {
	Metadata types.CueListMetadata `msgpack:"metadata" json:"metadata"`
}

func (p *Cueing) UpdateCueListLabel(sub string, request *UpdateCueListLabelRequest) (*UpdateCueListLabelResponse, error) {
	c := p.model.CueLists.Get(request.Number)
	if c == nil {
		return nil, &messaging.FriendlyError{FriendlyErr: CueListNotFound}
	}
	(*c).Metadata.Label = request.Label

	err := messaging.Publish(p.Messenger(), CueListMetadataUpdatedEventSubject, &CueListCreatedEvent{
		CueListMetadata: (*c).Metadata,
	})
	if err != nil {
		p.Logger().Error("Failed to publish updated cue list label", "error", err)
		return nil, err
	}

	return &UpdateCueListLabelResponse{Metadata: (*c).Metadata}, nil
}

const RenumberCueListRequestSubject = "request.cueing.cuelists.renumber"
const RenumberCueListEventSubject = "event.cueing.cuelists.renumber"

type RenumberCueListsRequest struct {
	OriginalNumber float64 `msgpack:"originalNumber" json:"originalNumber" validate:"required,gt=0"`
	NewNumber      float64 `msgpack:"newNumber" json:"newNumber" validate:"required,gt=0"`
}
type RenumberCueListsResponse struct{}

type RenumberCueListEvent struct {
	OriginalNumber float64 `msgpack:"originalNumber" json:"originalNumber"`
	NewNumber      float64 `msgpack:"newNumber" json:"newNumber"`
}

func (p *Cueing) RenumberCueList(sub string, request *RenumberCueListsRequest) (*RenumberCueListsResponse, error) {
	err := p.model.CueLists.Move(request.OriginalNumber, request.NewNumber)
	if errors.Is(err, util.ErrNotFound) {
		p.Logger().Debug("could not renumber cue list because the original cue list does not exist", "err", err, "originalNumber", request.OriginalNumber, "newNumber", request.NewNumber)
		return nil, &messaging.FriendlyError{FriendlyErr: CueListNotFound}
	}

	if errors.Is(err, util.ErrExists) {
		p.Logger().Debug("could not renumber cue list because the new number already exists", "err", err, "originalNumber", request.OriginalNumber, "newNumber", request.NewNumber)
		return nil, &messaging.FriendlyError{FriendlyErr: CueListNumberExists}
	}

	if err != nil {
		p.Logger().Error("Failed to renumber cue list", "err", err, "originalNumber", request.OriginalNumber, "newNumber", request.NewNumber)
		return nil, err
	}

	err = messaging.Publish(p.Messenger(), RenumberCueListEventSubject, &RenumberCueListEvent{
		OriginalNumber: request.OriginalNumber,
		NewNumber:      request.NewNumber,
	})
	if err != nil {
		p.Logger().Error("Failed to publish renumber cue list event", "err", err, "originalNumber", request.OriginalNumber, "newNumber", request.NewNumber)
		return nil, err
	}

	return &RenumberCueListsResponse{}, nil
}

const DeleteCueListRequestSubject = "request.cueing.cuelists.delete"
const DeleteCueListEventSubject = "event.cueing.cuelists.deleted"

type DeleteCueListsRequest struct {
	Number float64 `msgpack:"number" json:"number" validate:"required,gt=0"`
}
type DeleteCueListsResponse struct{}
type CueListDeletedEvent struct {
	Number float64 `msgpack:"number" json:"number"`
}

func (p *Cueing) DeleteCueList(sub string, request *DeleteCueListsRequest) (*DeleteCueListsResponse, error) {
	p.model.CueLists.Remove(request.Number)

	err := messaging.Publish(p.Messenger(), DeleteCueListEventSubject, &CueListDeletedEvent{})
	if err != nil {
		p.Logger().Error("failed to publish cue list deleted event", "error", err, "number", request.Number)
		return nil, err
	}

	return &DeleteCueListsResponse{}, nil
}
