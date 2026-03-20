// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package cues

const (
	RequestCreateCueList    = "request.cuelist.create"
	RequestUpdateCueList    = "request.cuelist.update"
	RequestGetCueList       = "request.cuelist.get"
	RequestEnumerateCueList = "request.cuelist.enumerate"
	RequestDeleteCueList    = "request.cuelist.delete"
	RequestMoveCueList      = "request.cuelist.move"

	EventNewCueList    = "event.cuelist.created"
	EventUpdateCueList = "event.cuelist.updated"
	EventDeleteCueList = "event.cuelist.deleted"
)

type CreateCueListInput struct {
	CueListNumber float64 `json:"cueListNumber" msgpack:"cueListNumber" validate:"gte=0"`
}

type CreateCueListOutput struct {
	CueListNumber float64 `json:"cueListNumber" msgpack:"cueListNumber" validate:"gte=0"`
}

type CueList struct {
	CueListNumber float64 `json:"cueListNumber" msgpack:"cueListNumber"`
	Label         string  `json:"label" msgpack:"label"`
	ListType      string  `json:"listType" msgpack:"listType"`
}

type NewCueListEvent struct {
	CueList CueList `json:"cueList" msgpack:"cueList"`
}

type UpdateCueListInput struct {
	CueListNumber float64 `json:"cueListNumber" msgpack:"cueListNumber" validate:"gt=0"`
	Key           string  `json:"key" msgpack:"key" validate:"required"`
	Value         string  `json:"value" msgpack:"value"`
}

type UpdateCueListOutput struct{}

type UpdateCueListEvent struct {
	CueList CueList `json:"cueList" msgpack:"cueList"`
}

type GetCueListInput struct {
	CueListNumber float64 `json:"cueListNumber" msgpack:"cueListNumber" validate:"gt=0"`
}

type GetCueListOutput struct {
	CueList CueList `json:"cueList" msgpack:"cueList"`
}

type EnumerateCueListInput struct{}

type EnumerateCueListOutput struct {
	CueLists []GetCueListOutput `json:"cueLists" msgpack:"cueLists"`
}

type DeleteCueListInput struct {
	CueListNumber float64 `json:"cueListNumber" msgpack:"cueListNumber" validate:"gt=0"`
}

type DeleteCueListOutput struct{}

type MoveCueListInput struct {
	OriginalCueListNumber float64 `json:"originalCueListNumber" msgpack:"originalCueListNumber" validate:"gt=0"`
	NewCueListNumber      float64 `json:"newCueListNumber" msgpack:"newCueListNumber" validate:"gt=0"`
}

type MoveCueListOutput struct {
	OriginalCueListNumber float64 `json:"originalCueListNumber" msgpack:"originalCueListNumber"`
	NewCueListNumber      float64 `json:"newCueListNumber" msgpack:"newCueListNumber"`
}

type DeleteCueListEvent struct {
	CueListNumber float64 `json:"cueListNumber" msgpack:"cueListNumber" validate:"gt=0"`
}
