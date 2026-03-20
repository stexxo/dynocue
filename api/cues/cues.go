// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package cues

const (
	RequestCreateCue    = "request.cue.create"
	RequestUpdateCue    = "request.cue.update"
	RequestGetCue       = "request.cue.get"
	RequestEnumerateCue = "request.cue.enumerate"
	RequestDeleteCue    = "request.cue.delete"
	RequestMoveCue      = "request.cue.move"

	EventNewCue    = "event.cue.created"
	EventUpdateCue = "event.cue.updated"
	EventDeleteCue = "event.cue.deleted"
)

type CreateCueInput struct {
	CueListNumber float64 `json:"cueListNumber" msgpack:"cueListNumber" validate:"gt=0"`
	CueNumber     float64 `json:"cueNumber" msgpack:"cueNumber" validate:"gte=0"`
}

type CreateCueOutput struct {
	CueListNumber float64 `json:"cueListNumber" msgpack:"cueListNumber"`
	CueNumber     float64 `json:"cueNumber" msgpack:"cueNumber"`
}

type Cue struct {
	CueNumber   float64 `json:"cueNumber" msgpack:"cueNumber"`
	Label       string  `json:"label" msgpack:"label"`
	Description string  `json:"description" msgpack:"description"`
}

type NewCueEvent struct {
	CueListNumber float64 `json:"cueListNumber" msgpack:"cueListNumber"`
	Cue           Cue     `json:"cue" msgpack:"cue"`
}

type UpdateCueInput struct {
	CueListNumber float64 `json:"cueListNumber" msgpack:"cueListNumber" validate:"gt=0"`
	CueNumber     float64 `json:"cueNumber" msgpack:"cueNumber" validate:"gt=0"`
	Key           string  `json:"key" msgpack:"key" validate:"required"`
	Value         string  `json:"value" msgpack:"value"`
}

type UpdateCueOutput struct{}

type UpdateCueEvent struct {
	CueListNumber float64 `json:"cueListNumber" msgpack:"cueListNumber"`
	Cue           Cue     `json:"cue" msgpack:"cue"`
}

type GetCueInput struct {
	CueListNumber float64 `json:"cueListNumber" msgpack:"cueListNumber" validate:"gt=0"`
	CueNumber     float64 `json:"cueNumber" msgpack:"cueNumber" validate:"gt=0"`
}

type GetCueOutput struct {
	CueListNumber float64 `json:"cueListNumber" msgpack:"cueListNumber"`
	Cue           Cue     `json:"cue" msgpack:"cue"`
}

type EnumerateCueInput struct {
	CueListNumber float64 `json:"cueListNumber" msgpack:"cueListNumber" validate:"gt=0"`
}

type EnumerateCueOutput struct {
	CueListNumber float64        `json:"cueListNumber" msgpack:"cueListNumber"`
	Cues          []GetCueOutput `json:"cues" msgpack:"cues"`
}

type DeleteCueInput struct {
	CueListNumber float64 `json:"cueListNumber" msgpack:"cueListNumber" validate:"gt=0"`
	CueNumber     float64 `json:"cueNumber" msgpack:"cueNumber" validate:"gt=0"`
}

type DeleteCueOutput struct{}

type DeleteCueEvent struct {
	CueListNumber float64 `json:"cueListNumber" msgpack:"cueListNumber"`
	CueNumber     float64 `json:"cueNumber" msgpack:"cueNumber"`
}

type MoveCueInput struct {
	CueListNumber     float64 `json:"cueListNumber" msgpack:"cueListNumber" validate:"gt=0"`
	OriginalCueNumber float64 `json:"originalCueNumber" msgpack:"originalCueNumber" validate:"gt=0"`
	NewCueNumber      float64 `json:"newCueNumber" msgpack:"newCueNumber" validate:"gt=0"`
}

type MoveCueOutput struct {
	CueListNumber float64 `json:"cueListNumber" msgpack:"cueListNumber"`
	NewCueNumber  float64 `json:"newCueNumber" msgpack:"newCueNumber"`
}
