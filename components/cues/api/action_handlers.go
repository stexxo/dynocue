package api

import (
	"errors"

	"github.com/stexxo/dynocue/components/cues/model"
	"github.com/stexxo/dynocue/components/cues/types"
	"github.com/stexxo/dynocue/core/messaging"
)

const (
	ActionNotFound     = "Action not found"
	ActionNumberExists = "Action number already exists"
)

func (c *CueingApi) registerActionApis() error {
	return errors.Join(
		messaging.Reply[CreateActionRequest, CreateActionResponse](c.messenger, true, CreateActionRequestSubject, c.CreateAction),
		messaging.Reply[EnumerateActionsRequest, EnumerateActionsResponse](c.messenger, true, EnumerateActionsRequestSubject, c.EnumerateActions),
		messaging.Reply[GetActionByIdRequest, GetActionByIdResponse](c.messenger, true, GetActionByIdRequestSubject, c.GetActionById),
		messaging.Reply[DeleteActionRequest, DeleteActionResponse](c.messenger, true, DeleteActionRequestSubject, c.DeleteAction),
		messaging.Reply[UpdateActionRequest, UpdateActionResponse](c.messenger, true, UpdateActionRequestSubject, c.UpdateAction),
		messaging.Reply[UpdateActionFieldRequest, UpdateActionFieldResponse](c.messenger, true, UpdateActionFieldRequestSubject, c.UpdateActionField),
	)
}

const CreateActionRequestSubject = "request.cueing.action.create"

type CreateActionRequest struct {
	CueId      string `msgpack:"cueId" json:"cueId" validate:"required"`
	TemplateId string `msgpack:"templateId" json:"templateId" validate:"required"`
	Number     uint   `msgpack:"number" json:"number" validate:"gte=0"`
}

type CreateActionResponse struct {
	ActionId string `msgpack:"actionId" json:"actionId"`
	Number   uint   `msgpack:"number" json:"number"`
}

func (c *CueingApi) CreateAction(sub string, request *CreateActionRequest) (*CreateActionResponse, error) {
	id, num, err := c.model.CreateAction(request.CueId, request.TemplateId, request.Number)
	if errors.Is(err, model.ErrNumberExists) {
		return nil, errors.Join(err, &messaging.FriendlyError{FriendlyErr: ActionNumberExists})
	}
	if err != nil {
		return nil, err
	}
	return &CreateActionResponse{ActionId: id, Number: num}, nil
}

const EnumerateActionsRequestSubject = "request.cueing.action.enumerate"

type EnumerateActionsRequest struct {
	CueId string `msgpack:"cueId" json:"cueId" validate:"required"`
}

type EnumerateActionsResponse struct {
	Actions []types.Action `msgpack:"actions" json:"actions"`
}

func (c *CueingApi) EnumerateActions(sub string, request *EnumerateActionsRequest) (*EnumerateActionsResponse, error) {
	actions, err := c.model.EnumerateActions(request.CueId)
	if err != nil {
		return nil, err
	}
	return &EnumerateActionsResponse{Actions: actions}, nil
}

const GetActionByIdRequestSubject = "request.cueing.action.get.id"

type GetActionByIdRequest struct {
	ActionId string `msgpack:"actionId" json:"actionId" validate:"required"`
}

type GetActionByIdResponse struct {
	Action types.Action `msgpack:"action" json:"action"`
}

func (c *CueingApi) GetActionById(sub string, request *GetActionByIdRequest) (*GetActionByIdResponse, error) {
	out, err := c.model.GetActionById(request.ActionId)
	if errors.Is(err, model.ErrActionNotFound) {
		return nil, errors.Join(err, &messaging.FriendlyError{FriendlyErr: ActionNotFound})
	}
	if err != nil {
		return nil, err
	}
	return &GetActionByIdResponse{
		Action: *out,
	}, nil
}

const DeleteActionRequestSubject = "request.cueing.action.delete"

type DeleteActionRequest struct {
	ActionId string `msgpack:"actionId" json:"actionId" validate:"required"`
}

type DeleteActionResponse struct{}

func (c *CueingApi) DeleteAction(sub string, request *DeleteActionRequest) (*DeleteActionResponse, error) {
	err := c.model.DeleteAction(request.ActionId)
	if err != nil {
		return nil, err
	}

	return &DeleteActionResponse{}, nil
}

const UpdateActionRequestSubject = "request.cueing.action.update"

type UpdateActionRequest struct {
	ActionId string `msgpack:"actionId" json:"actionId" validate:"required"`
	Field    string `msgpack:"field" json:"field" validate:"required,ne=actionId,ne=number,ne=cueId"`
	Value    any    `msgpack:"value" json:"value" validate:"required"`
}

type UpdateActionResponse struct{}

func (c *CueingApi) UpdateAction(sub string, request *UpdateActionRequest) (*UpdateActionResponse, error) {
	err := c.model.UpdateAction(request.ActionId, request.Field, request.Value)
	if errors.Is(err, model.ErrActionNotFound) {
		return nil, errors.Join(err, &messaging.FriendlyError{FriendlyErr: ActionNotFound})
	}
	if err != nil {
		return nil, err
	}

	return &UpdateActionResponse{}, nil
}

const UpdateActionFieldRequestSubject = "request.cueing.action.field.update"

type UpdateActionFieldRequest struct {
	ActionId  string `msgpack:"actionId" json:"actionId" validate:"required"`
	FieldName string `msgpack:"fieldName" json:"fieldName" validate:"required"`
	Value     any    `msgpack:"value" json:"value" validate:"required"`
}

type UpdateActionFieldResponse struct{}

func (c *CueingApi) UpdateActionField(sub string, request *UpdateActionFieldRequest) (*UpdateActionFieldResponse, error) {
	err := c.model.UpdateActionField(request.ActionId, request.FieldName, request.Value)
	if errors.Is(err, model.ErrActionNotFound) {
		return nil, errors.Join(err, &messaging.FriendlyError{FriendlyErr: ActionNotFound})
	}
	if err != nil {
		return nil, err
	}

	return &UpdateActionFieldResponse{}, nil
}
