// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cues

import (
	"errors"
	"fmt"

	"github.com/stexxo/dynocue/components/cues/types"
	"github.com/stexxo/dynocue/core/messaging"
	"github.com/stexxo/dynocue/util"
)

const ActionNotFound = "Action Not Found"
const ActionNumberExists = "Action Number Already Exists"

// EnumerateActions

const EnumerateActionsRequestSubject = "request.cueing.actions.enumerate"

type EnumerateActionsRequest struct {
	CueListId string `msgpack:"cueListId" json:"cueListId" validate:"required"`
	CueId     string `msgpack:"cueId" json:"cueId" validate:"required"`
}

type EnumerateActionsResponse struct {
	Actions []types.Action `msgpack:"actions" json:"actions"`
}

func (p *Cueing) EnumerateActions(sub string, request *EnumerateActionsRequest) (*EnumerateActionsResponse, error) {
	cl, err := p.getCueListById(request.CueListId)
	if err != nil {
		return nil, err
	}

	cue, err := p.getCueById(cl, request.CueId)
	if err != nil {
		return nil, err
	}

	var actions []types.Action
	cue.Actions.ForEach(func(action *types.Action) {
		actions = append(actions, *action)
	})

	return &EnumerateActionsResponse{Actions: actions}, nil
}

// GetActionByNumber

const GetActionByNumberRequestSubject = "request.cueing.actions.get.number"

type GetActionByNumberRequest struct {
	CueListId    string  `msgpack:"cueListId" json:"cueListId" validate:"required"`
	CueId        string  `msgpack:"cueId" json:"cueId" validate:"required"`
	ActionNumber float64 `msgpack:"actionNumber" json:"actionNumber" validate:"required"`
}

type GetActionByNumberResponse struct {
	Action types.Action `msgpack:"action" json:"action"`
}

func (p *Cueing) GetActionByNumber(sub string, request *GetActionByNumberRequest) (*GetActionByNumberResponse, error) {
	cl, err := p.getCueListById(request.CueListId)
	if err != nil {
		return nil, err
	}

	cue, err := p.getCueById(cl, request.CueId)
	if err != nil {
		return nil, err
	}

	action := cue.Actions.GetByNumber(request.ActionNumber)
	if action == nil {
		return nil, &messaging.FriendlyError{FriendlyErr: ActionNotFound}
	}

	return &GetActionByNumberResponse{Action: **action}, nil
}

// GetActionById

const GetActionByIdRequestSubject = "request.cueing.actions.get.id"

type GetActionByIdRequest struct {
	CueListId string `msgpack:"cueListId" json:"cueListId" validate:"required"`
	CueId     string `msgpack:"cueId" json:"cueId" validate:"required"`
	ActionId  string `msgpack:"actionId" json:"actionId" validate:"required"`
}

type GetActionByIdResponse struct {
	Action types.Action `msgpack:"action" json:"action"`
}

func (p *Cueing) GetActionById(sub string, request *GetActionByIdRequest) (*GetActionByIdResponse, error) {
	cl, err := p.getCueListById(request.CueListId)
	if err != nil {
		return nil, err
	}

	cue, err := p.getCueById(cl, request.CueId)
	if err != nil {
		return nil, err
	}

	action := cue.Actions.GetFunc(func(a *types.Action) bool {
		return a.Id == request.ActionId
	})
	if action == nil {
		return nil, &messaging.FriendlyError{FriendlyErr: ActionNotFound}
	}

	return &GetActionByIdResponse{Action: **action}, nil
}

// DeleteAction

const DeleteActionRequestSubject = "request.cueing.actions.delete"
const ActionDeletedEventSubject = "event.cueing.actions.deleted"

type DeleteActionRequest struct {
	CueListId string `msgpack:"cueListId" json:"cueListId" validate:"required"`
	CueId     string `msgpack:"cueId" json:"cueId" validate:"required"`
	ActionId  string `msgpack:"actionId" json:"actionId" validate:"required"`
}

type DeleteActionResponse struct{}

type ActionDeletedEvent struct {
	CueListId string `msgpack:"cueListId" json:"cueListId"`
	CueId     string `msgpack:"cueId" json:"cueId"`
	ActionId  string `msgpack:"actionId" json:"actionId"`
}

func (p *Cueing) DeleteAction(sub string, request *DeleteActionRequest) (*DeleteActionResponse, error) {
	cl, err := p.getCueListById(request.CueListId)
	if err != nil {
		return nil, err
	}

	cue, err := p.getCueById(cl, request.CueId)
	if err != nil {
		return nil, err
	}

	cue.Actions.RemoveFunc(func(a *types.Action) bool {
		return a.Id == request.ActionId
	})

	err = messaging.Publish(p.Messenger(), ActionDeletedEventSubject, &ActionDeletedEvent{
		CueListId: request.CueListId,
		CueId:     request.CueId,
		ActionId:  request.ActionId,
	})
	if err != nil {
		return nil, err
	}

	return &DeleteActionResponse{}, nil
}

// RenumberAction

const RenumberActionRequestSubject = "request.cueing.actions.renumber"
const ActionRenumberedEventSubject = "event.cueing.actions.renumbered"

type RenumberActionRequest struct {
	CueListId string  `msgpack:"cueListId" json:"cueListId" validate:"required"`
	CueId     string  `msgpack:"cueId" json:"cueId" validate:"required"`
	ActionId  string  `msgpack:"actionId" json:"actionId" validate:"required"`
	NewNumber float64 `msgpack:"newNumber" json:"newNumber" validate:"required"`
}

type RenumberActionResponse struct{}

type ActionRenumberedEvent struct {
	CueListId string  `msgpack:"cueListId" json:"cueListId"`
	CueId     string  `msgpack:"cueId" json:"cueId"`
	ActionId  string  `msgpack:"actionId" json:"actionId"`
	NewNumber float64 `msgpack:"newNumber" json:"newNumber"`
}

func (p *Cueing) RenumberAction(sub string, request *RenumberActionRequest) (*RenumberActionResponse, error) {
	cl, err := p.getCueListById(request.CueListId)
	if err != nil {
		return nil, err
	}

	cue, err := p.getCueById(cl, request.CueId)
	if err != nil {
		return nil, err
	}

	err = cue.Actions.MoveFunc(func(a *types.Action) bool {
		return a.Id == request.ActionId
	}, request.NewNumber)
	if errors.Is(err, util.ErrNotFound) {
		return nil, &messaging.FriendlyError{FriendlyErr: ActionNotFound}
	}
	if errors.Is(err, util.ErrExists) {
		return nil, &messaging.FriendlyError{FriendlyErr: ActionNumberExists}
	}
	if err != nil {
		return nil, err
	}

	err = messaging.Publish(p.Messenger(), ActionRenumberedEventSubject, &ActionRenumberedEvent{
		CueListId: request.CueListId,
		CueId:     request.CueId,
		ActionId:  request.ActionId,
		NewNumber: request.NewNumber,
	})
	if err != nil {
		return nil, err
	}

	return &RenumberActionResponse{}, nil
}

// UpdateAction

const UpdateActionRequestSubject = "request.cueing.actions.update"
const ActionUpdatedEventSubject = "event.cueing.actions.updated"

type UpdateActionRequest struct {
	CueListId string      `msgpack:"cueListId" json:"cueListId" validate:"required"`
	CueId     string      `msgpack:"cueId" json:"cueId" validate:"required"`
	ActionId  string      `msgpack:"actionId" json:"actionId" validate:"required"`
	Field     string      `msgpack:"field" json:"field" validate:"required,oneof=subject delay follow"`
	Value     interface{} `msgpack:"value" json:"value"`
}

type UpdateActionResponse struct {
	Action types.Action `msgpack:"action" json:"action"`
}

type ActionUpdatedEvent struct {
	CueListId string       `msgpack:"cueListId" json:"cueListId"`
	CueId     string       `msgpack:"cueId" json:"cueId"`
	Action    types.Action `msgpack:"action" json:"action"`
}

func (p *Cueing) UpdateAction(sub string, request *UpdateActionRequest) (*UpdateActionResponse, error) {
	cl, err := p.getCueListById(request.CueListId)
	if err != nil {
		return nil, err
	}

	cue, err := p.getCueById(cl, request.CueId)
	if err != nil {
		return nil, err
	}

	action := cue.Actions.GetFunc(func(a *types.Action) bool {
		return a.Id == request.ActionId
	})
	if action == nil {
		return nil, &messaging.FriendlyError{FriendlyErr: ActionNotFound}
	}

	err = util.UpdateStructByTag("json", request.Field, request.Value, *action)
	if err != nil {
		return nil, err
	}

	err = messaging.Publish(p.Messenger(), ActionUpdatedEventSubject, &ActionUpdatedEvent{
		CueListId: request.CueListId,
		CueId:     request.CueId,
		Action:    **action,
	})
	if err != nil {
		return nil, err
	}

	return &UpdateActionResponse{Action: **action}, nil
}

// UpdateActionField

const UpdateActionFieldRequestSubject = "request.cueing.actions.field.update"

type UpdateActionFieldRequest struct {
	CueListId string      `msgpack:"cueListId" json:"cueListId" validate:"required"`
	CueId     string      `msgpack:"cueId" json:"cueId" validate:"required"`
	ActionId  string      `msgpack:"actionId" json:"actionId" validate:"required"`
	FieldName string      `msgpack:"fieldName" json:"fieldName" validate:"required"`
	Value     interface{} `msgpack:"value" json:"value"`
}

type UpdateActionFieldResponse struct {
	Action types.Action `msgpack:"action" json:"action"`
}

func (p *Cueing) UpdateActionField(sub string, request *UpdateActionFieldRequest) (*UpdateActionFieldResponse, error) {
	cl, err := p.getCueListById(request.CueListId)
	if err != nil {
		return nil, err
	}

	cue, err := p.getCueById(cl, request.CueId)
	if err != nil {
		return nil, err
	}

	action := cue.Actions.GetFunc(func(a *types.Action) bool {
		return a.Id == request.ActionId
	})
	if action == nil {
		return nil, &messaging.FriendlyError{FriendlyErr: ActionNotFound}
	}

	foundField := false
	for i := range (*action).Fields {
		if (*action).Fields[i].FieldName == request.FieldName {
			(*action).Fields[i].Value = request.Value
			foundField = true
			break
		}
	}

	if !foundField {
		return nil, fmt.Errorf("field %s not found in action %s", request.FieldName, request.ActionId)
	}

	err = messaging.Publish(p.Messenger(), ActionUpdatedEventSubject, &ActionUpdatedEvent{
		CueListId: request.CueListId,
		CueId:     request.CueId,
		Action:    **action,
	})
	if err != nil {
		return nil, err
	}

	return &UpdateActionFieldResponse{Action: **action}, nil
}
