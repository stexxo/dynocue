// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cues

const CreateCueListRequestSubject = "request.cueing.cuelists.create"
const CueListCreatedEventSubject = "event.cueing.cuelists.created"

type CreateCueListRequest struct{}

type CreateCueListResponse struct{}

type CueListCreatedEvent struct{}

func (p *Cueing) CreateCueList(sub string, request CreateCueListRequest) (*CreateCueListResponse, error) {
	return &CreateCueListResponse{}, nil
}
