// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cues

import (
	"slices"

	"github.com/stexxo/dynocue/components/cues/types"
	"github.com/stexxo/dynocue/core/messaging"
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
	for c := range slices.Values(p.model.CueLists.Data) {
		out = append(out, c.Metadata)
	}

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
	Label  string  `msgpack:"label" json:"label" validate:"required"`
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
		return nil, err
	}

	return &UpdateCueListLabelResponse{Metadata: (*c).Metadata}, nil
}
