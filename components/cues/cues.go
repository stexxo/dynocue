// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cues

import (
	"github.com/stexxo/dynocue/components/cues/types"
	"github.com/stexxo/dynocue/core/messaging"
)

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
	cue := types.NewCue(req.CueNumber)

	cl := p.model.CueLists.GetFunc(func(list *types.CueList) bool {
		return list.Id() == req.CueListId
	})

	if cl == nil {
		return nil, &messaging.FriendlyError{FriendlyErr: CueListNotFound}
	}

	(*cl).Cues.Add(cue)

	err := messaging.Publish(p.Messenger(), CueCreatedEventSubject, &CueCreatedEvent{
		CueListId: req.CueListId,
		Metadata:  cue.Metadata,
	})
	if err != nil {
		p.Logger().Error("failed to publish cue created event", "error", err)
		return nil, err
	}

	return &CreateCueResponse{CueListId: req.CueListId, CueId: cue.Metadata.Id, CueNumber: cue.Metadata.Number}, nil
}
