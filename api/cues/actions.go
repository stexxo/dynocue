// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package cues

const (
	RequestCreateAction    = "request.action.create"
	RequestUpdateAction    = "request.action.update"
	RequestGetAction       = "request.action.get"
	RequestEnumerateAction = "request.action.enumerate"
	RequestDeleteAction    = "request.action.delete"
	RequestMoveAction      = "request.action.move"

	EventNewAction    = "event.action.created"
	EventUpdateAction = "event.action.updated"
	EventDeleteAction = "event.action.deleted"
	EventMoveAction   = "event.action.moved"
)

type CreateActionInput struct {
	CueListNumber float64 `json:"cueListNumber" msgpack:"cueListNumber" validate:"gt=0"`
	CueNumber     float64 `json:"cueNumber" msgpack:"cueNumber" validate:"gt=0"`
	ActionNumber  float64 `json:"actionNumber" msgpack:"actionNumber" validate:"gte=0"`
}

type CreateActionOutput struct {
	CueListNumber float64 `json:"cueListNumber" msgpack:"cueListNumber"`
	CueNumber     float64 `json:"cueNumber" msgpack:"cueNumber"`
	ActionNumber  float64 `json:"actionNumber" msgpack:"actionNumber"`
}

type CueAction struct {
	ActionNumber float64 `json:"actionNumber" msgpack:"actionNumber"`
	Label        string  `json:"label" msgpack:"label"`
	SourceType   string  `json:"sourceType" msgpack:"sourceType"`
	Action       string  `json:"action" msgpack:"action"`
	Target       float64 `json:"target" msgpack:"target"`
}

type NewActionEvent struct {
	CueListNumber float64   `json:"cueListNumber" msgpack:"cueListNumber"`
	CueNumber     float64   `json:"cueNumber" msgpack:"cueNumber"`
	Action        CueAction `json:"action" msgpack:"action"`
}

type UpdateActionInput struct {
	CueListNumber float64 `json:"cueListNumber" msgpack:"cueListNumber" validate:"gt=0"`
	CueNumber     float64 `json:"cueNumber" msgpack:"cueNumber" validate:"gt=0"`
	ActionNumber  float64 `json:"actionNumber" msgpack:"actionNumber" validate:"gt=0"`
	Key           string  `json:"key" msgpack:"key" validate:"required"`
	Value         string  `json:"value" msgpack:"value"`
}

type UpdateActionOutput struct{}

type UpdateActionEvent struct {
	CueListNumber float64   `json:"cueListNumber" msgpack:"cueListNumber"`
	CueNumber     float64   `json:"cueNumber" msgpack:"cueNumber"`
	Action        CueAction `json:"action" msgpack:"action"`
}

type GetActionInput struct {
	CueListNumber float64 `json:"cueListNumber" msgpack:"cueListNumber" validate:"gt=0"`
	CueNumber     float64 `json:"cueNumber" msgpack:"cueNumber" validate:"gt=0"`
	ActionNumber  float64 `json:"actionNumber" msgpack:"actionNumber" validate:"gt=0"`
}

type GetActionOutput struct {
	CueListNumber float64   `json:"cueListNumber" msgpack:"cueListNumber"`
	CueNumber     float64   `json:"cueNumber" msgpack:"cueNumber"`
	Action        CueAction `json:"action" msgpack:"action"`
}

type EnumerateActionInput struct {
	CueListNumber float64 `json:"cueListNumber" msgpack:"cueListNumber" validate:"gt=0"`
	CueNumber     float64 `json:"cueNumber" msgpack:"cueNumber" validate:"gt=0"`
}

type EnumerateActionOutput struct {
	CueListNumber float64           `json:"cueListNumber" msgpack:"cueListNumber"`
	CueNumber     float64           `json:"cueNumber" msgpack:"cueNumber"`
	Actions       []GetActionOutput `json:"actions" msgpack:"actions"`
}

type DeleteActionInput struct {
	CueListNumber float64 `json:"cueListNumber" msgpack:"cueListNumber" validate:"gt=0"`
	CueNumber     float64 `json:"cueNumber" msgpack:"cueNumber" validate:"gt=0"`
	ActionNumber  float64 `json:"actionNumber" msgpack:"actionNumber" validate:"gt=0"`
}

type DeleteActionOutput struct{}

type DeleteActionEvent struct {
	CueListNumber float64 `json:"cueListNumber" msgpack:"cueListNumber"`
	CueNumber     float64 `json:"cueNumber" msgpack:"cueNumber"`
	ActionNumber  float64 `json:"actionNumber" msgpack:"actionNumber"`
}

type MoveActionInput struct {
	CueListNumber        float64 `json:"cueListNumber" msgpack:"cueListNumber" validate:"gt=0"`
	CueNumber            float64 `json:"cueNumber" msgpack:"cueNumber" validate:"gt=0"`
	OriginalActionNumber float64 `json:"originalActionNumber" msgpack:"originalActionNumber" validate:"gt=0"`
	NewActionNumber      float64 `json:"newActionNumber" msgpack:"newActionNumber" validate:"gt=0"`
}

type MoveActionOutput struct {
	CueListNumber   float64 `json:"cueListNumber" msgpack:"cueListNumber"`
	CueNumber       float64 `json:"cueNumber" msgpack:"cueNumber"`
	NewActionNumber float64 `json:"newActionNumber" msgpack:"newActionNumber"`
}

type MoveActionEvent struct {
	CueListNumber        float64 `json:"cueListNumber" msgpack:"cueListNumber"`
	CueNumber            float64 `json:"cueNumber" msgpack:"cueNumber"`
	OriginalActionNumber float64 `json:"originalActionNumber" msgpack:"originalActionNumber"`
	NewActionNumber      float64 `json:"newActionNumber" msgpack:"newActionNumber"`
}
