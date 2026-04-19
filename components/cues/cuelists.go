// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cues

import "github.com/stexxo/dynocue/core/messaging"

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
	Number      float64 `msgpack:"number" json:"number" validate:"required, gt=0"`
	Label       string  `msgpack:"label" json:"label"`
	CueListType string  `msgpack:"cueListType" json:"cueListType" validate:"required, oneOf=SEQUENTIAL"`
}

func (p *Cueing) CreateCueList(sub string, request *CreateCueListRequest) (*CreateCueListResponse, error) {
	c := p.model.CueLists.create(request.Number, request.CueListType)
	if c == nil {
		return nil, &messaging.FriendlyError{FriendlyErr: "cuelist number already exists"}
	}

	err := messaging.Publish(p.Messenger(), CueListCreatedEventSubject, &CueListCreatedEvent{
		Number:      c.Number,
		Label:       request.CueListType,
		CueListType: request.CueListType,
	})
	if err != nil {
		return nil, err
	}

	return &CreateCueListResponse{
		Number: c.Number,
	}, nil
}

const EnumerateCueListsRequestSubject = "request.cueing.cuelists.enumerate"

type EnumerateCueListsRequest struct{}
type EnumerateCueListsResponse struct {
	CueLists []CueListEnumeration `msgpack:"cueLists" json:"cueLists"`
}
type CueListEnumeration struct {
	Number      float64 `msgpack:"number" json:"number"`
	Label       string  `msgpack:"label" json:"label"`
	CueListType string  `msgpack:"cueListType" json:"cueListType"`
}

func (p *Cueing) EnumerateCueLists(sub string, request *EnumerateCueListsRequest) (*EnumerateCueListsResponse, error) {
	out := make([]CueListEnumeration, 0, p.model.CueLists.Len())
	for _, c := range p.model.CueLists {
		out = append(out, CueListEnumeration{
			Number:      c.Number,
			Label:       c.Label,
			CueListType: c.CueListType,
		})
	}

	return &EnumerateCueListsResponse{CueLists: out}, nil
}
