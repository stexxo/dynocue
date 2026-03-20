// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package cues

const (
	RequestCreateCue         = "request.cue.create"
	RequestUpdateCueMetadata = "request.cue.metadata.update"
	RequestGetCueMetadata    = "request.cue.metadata.get"
	RequestEnumerateCue      = "request.cue.enumerate"
	RequestDeleteCue         = "request.cue.delete"
	RequestMoveCue           = "request.cue.move"

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

type NewCueEvent struct {
	CueListNumber float64 `json:"cueListNumber" msgpack:"cueListNumber"`
	CueNumber     float64 `json:"cueNumber" msgpack:"cueNumber"`
	Label         string  `json:"label" msgpack:"label"`
}

type UpdateCueMetadataInput struct {
	CueListNumber float64 `json:"cueListNumber" msgpack:"cueListNumber" validate:"gt=0"`
	CueNumber     float64 `json:"cueNumber" msgpack:"cueNumber" validate:"gt=0"`
	Key           string  `json:"key" msgpack:"key" validate:"required"`
	Value         string  `json:"value" msgpack:"value"`
}

type UpdateCueMetadataOutput struct{}

type UpdateCueMetadataEvent struct {
	CueListNumber float64 `json:"cueListNumber" msgpack:"cueListNumber"`
	CueNumber     float64 `json:"cueNumber" msgpack:"cueNumber"`
	Label         string  `json:"label" msgpack:"label"`
}

type GetCueMetadataInput struct {
	CueListNumber float64 `json:"cueListNumber" msgpack:"cueListNumber" validate:"gt=0"`
	CueNumber     float64 `json:"cueNumber" msgpack:"cueNumber" validate:"gt=0"`
}

type GetCueMetadataOutput struct {
	CueListNumber float64 `json:"cueListNumber" msgpack:"cueListNumber"`
	CueNumber     float64 `json:"cueNumber" msgpack:"cueNumber"`
	Label         string  `json:"label" msgpack:"label"`
}

type EnumerateCueInput struct {
	CueListNumber float64 `json:"cueListNumber" msgpack:"cueListNumber" validate:"gt=0"`
}

type EnumerateCueOutput struct {
	CueListNumber float64                `json:"cueListNumber" msgpack:"cueListNumber"`
	Cues          []GetCueMetadataOutput `json:"cues" msgpack:"cues"`
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
