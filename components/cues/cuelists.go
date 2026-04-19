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
	Number      float64 `msgpack:"number" json:"number" validate:"required, gte=0"`
	CueListType string  `msgpack:"cueListType" json:"cueListType" validate:"required, oneOf=SEQUENTIAL"`
}

type CreateCueListResponse struct {
	Number float64 `msgpack:"number" json:"number"`
}

type CueListCreatedEvent struct {
	CueListMetadata types.CueListMetadata `msgpack:"cueListMetadata" json:"cueListMetadata"`
}

const CueListNumberExists = "Cue List Number Already Exists"

func (p *Cueing) CreateCueList(sub string, request *CreateCueListRequest) (*CreateCueListResponse, error) {
	cPtr := p.model.CueLists.Add(request.Number)
	if cPtr == nil {
		return nil, &messaging.FriendlyError{FriendlyErr: CueListNumberExists}
	}
	c := *cPtr

	c.Metadata.Label = request.CueListType

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
	Number float64 `msgpack:"number" json:"number" validate:"required, gt=0"`
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
