// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cues

import (
	"fmt"
	"slices"

	"github.com/stexxo/dynocue/components/cues/types"
	"github.com/stexxo/dynocue/core/messaging"
	"github.com/stexxo/dynocue/util"
)

const ActionNotFound = "Action Not Found"
const ActionNumberExists = "Action Number Already Exists"

// CreateAction

const CreateActionRequestSubject = "request.cueing.actions.create"
const ActionCreatedEventSubject = "event.cueing.actions.created"

type CreateActionRequest struct {
	CueListId  string `msgpack:"cueListId" json:"cueListId" validate:"required"`
	CueId      string `msgpack:"cueId" json:"cueId" validate:"required"`
	TemplateId string `msgpack:"templateId" json:"templateId" validate:"required"`
}

type CreateActionResponse struct {
	Action types.Action `msgpack:"action" json:"action"`
}

type ActionCreatedEvent struct {
	CueListId string       `msgpack:"cueListId" json:"cueListId"`
	CueId     string       `msgpack:"cueId" json:"cueId"`
	Action    types.Action `msgpack:"action" json:"action"`
}

func (p *Cueing) CreateAction(sub string, request *CreateActionRequest) (*CreateActionResponse, error) {
	cue, err := p.getCueById(request.CueListId, request.CueId)
	if err != nil {
		return nil, err
	}

	template := p.actionTemplates.GetTemplateById(request.TemplateId)
	if template == nil {
		return nil, &messaging.FriendlyError{FriendlyErr: ActionTemplateNotFound}
	}

	action := types.NewActionByTemplate(template)

	cue.Actions = append(cue.Actions, *action)

	err = messaging.Publish(p.Messenger(), ActionCreatedEventSubject, &ActionCreatedEvent{
		CueListId: request.CueListId,
		CueId:     request.CueId,
		Action:    *action,
	})
	if err != nil {
		p.Logger().Error("failed to publish action created event", "error", err)
		return nil, err
	}

	return &CreateActionResponse{Action: *action}, nil
}

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
	cue, err := p.getCueById(request.CueListId, request.CueId)
	if err != nil {
		return nil, err
	}

	return &EnumerateActionsResponse{Actions: cue.Actions}, nil
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
	action, err := p.getActionById(request.CueListId, request.CueId, request.ActionId)
	if err != nil {
		return nil, err
	}

	return &GetActionByIdResponse{Action: *action}, nil
}

func (p *Cueing) getActionById(cueListId string, cueId string, actionId string) (*types.Action, error) {
	cue, err := p.getCueById(cueListId, cueId)
	if err != nil {
		return nil, err
	}

	actionIdx := slices.IndexFunc(cue.Actions, func(a types.Action) bool {
		return a.Id == actionId
	})
	if actionIdx == -1 {
		return nil, &messaging.FriendlyError{FriendlyErr: ActionNotFound}
	}

	return &cue.Actions[actionIdx], nil
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
	cue, _ := p.getCueById(request.CueListId, request.CueId)

	slices.DeleteFunc(cue.Actions, func(a types.Action) bool {
		return a.Id == request.ActionId
	})

	err := messaging.Publish(p.Messenger(), ActionDeletedEventSubject, &ActionDeletedEvent{
		CueListId: request.CueListId,
		CueId:     request.CueId,
		ActionId:  request.ActionId,
	})
	if err != nil {
		return nil, err
	}

	return &DeleteActionResponse{}, nil
}

// UpdateAction

const UpdateActionRequestSubject = "request.cueing.actions.update"
const ActionUpdatedEventSubject = "event.cueing.actions.updated"

type UpdateActionRequest struct {
	CueListId string      `msgpack:"cueListId" json:"cueListId" validate:"required"`
	CueId     string      `msgpack:"cueId" json:"cueId" validate:"required"`
	ActionId  string      `msgpack:"actionId" json:"actionId" validate:"required"`
	Field     string      `msgpack:"field" json:"field" validate:"required,ne=id,ne=subject,ne=templateId,ne=fields"`
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
	action, err := p.getActionById(request.CueListId, request.CueId, request.ActionId)
	if err != nil {
		return nil, err
	}

	err = util.UpdateStructByTag("json", request.Field, request.Value, action)
	if err != nil {
		return nil, err
	}

	err = messaging.Publish(p.Messenger(), ActionUpdatedEventSubject, &ActionUpdatedEvent{
		CueListId: request.CueListId,
		CueId:     request.CueId,
		Action:    *action,
	})
	if err != nil {
		return nil, err
	}

	return &UpdateActionResponse{Action: *action}, nil
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
	action, err := p.getActionById(request.CueListId, request.CueId, request.ActionId)
	if err != nil {
		return nil, err
	}

	foundField := false
	for i := range action.Fields {
		if action.Fields[i].FieldName == request.FieldName {
			action.Fields[i].Value = request.Value
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
		Action:    *action,
	})
	if err != nil {
		return nil, err
	}

	return &UpdateActionFieldResponse{Action: *action}, nil
}
