// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package api

import (
	"errors"

	"github.com/stexxo/dynocue/components/cues/model"
	"github.com/stexxo/dynocue/components/cues/types"
	"github.com/stexxo/dynocue/core/messaging"
)

const (
	ActionTemplateNotFound = "Action template not found"
	ActionTemplateExists   = "Action template already exists"
)

func (c *CueingApi) registerActionTemplateApis() error {
	return errors.Join(
		messaging.Reply[RegisterActionTemplateRequest, RegisterActionTemplateResponse](c.messenger, true, RegisterActionTemplateRequestSubject, c.RegisterActionTemplate),
		messaging.Reply[EnumerateActionTemplatesRequest, EnumerateActionTemplatesResponse](c.messenger, true, EnumerateActionTemplatesRequestSubject, c.EnumerateActionTemplates),
		messaging.Reply[GetActionTemplateByIdRequest, GetActionTemplateByIdResponse](c.messenger, true, GetActionTemplateByIdRequestSubject, c.GetActionTemplateById),
		messaging.Reply[DeleteActionTemplateRequest, DeleteActionTemplateResponse](c.messenger, true, DeleteActionTemplateRequestSubject, c.DeleteActionTemplate),
	)
}

const RegisterActionTemplateRequestSubject = "request.cueing.actiontemplate.register"

type RegisterActionTemplateRequest struct {
	Template types.ActionTemplate `msgpack:"template" json:"template" validate:"required"`
}

type RegisterActionTemplateResponse struct {
	TemplateId string `msgpack:"templateId" json:"templateId"`
}

func (c *CueingApi) RegisterActionTemplate(sub string, request *RegisterActionTemplateRequest) (*RegisterActionTemplateResponse, error) {
	err := c.model.RegisterActionTemplate(&request.Template)
	if errors.Is(err, model.ErrActionTemplateExists) {
		return nil, errors.Join(err, &messaging.FriendlyError{FriendlyErr: ActionTemplateExists})
	}
	if err != nil {
		return nil, err
	}
	return &RegisterActionTemplateResponse{TemplateId: request.Template.TemplateId}, nil
}

const EnumerateActionTemplatesRequestSubject = "request.cueing.actiontemplate.enumerate"

type EnumerateActionTemplatesRequest struct{}

type EnumerateActionTemplatesResponse struct {
	Templates []types.ActionTemplate `msgpack:"templates" json:"templates"`
}

func (c *CueingApi) EnumerateActionTemplates(sub string, request *EnumerateActionTemplatesRequest) (*EnumerateActionTemplatesResponse, error) {
	templates, err := c.model.EnumerateActionTemplates()
	if err != nil {
		return nil, err
	}
	return &EnumerateActionTemplatesResponse{Templates: templates}, nil
}

const GetActionTemplateByIdRequestSubject = "request.cueing.actiontemplate.get.id"

type GetActionTemplateByIdRequest struct {
	TemplateId string `msgpack:"templateId" json:"templateId" validate:"required"`
}

type GetActionTemplateByIdResponse struct {
	Template types.ActionTemplate `msgpack:"template" json:"template"`
}

func (c *CueingApi) GetActionTemplateById(sub string, request *GetActionTemplateByIdRequest) (*GetActionTemplateByIdResponse, error) {
	out, err := c.model.GetActionTemplateById(request.TemplateId)
	if errors.Is(err, model.ErrActionTemplateNotFound) {
		return nil, errors.Join(err, &messaging.FriendlyError{FriendlyErr: ActionTemplateNotFound})
	}
	if err != nil {
		return nil, err
	}
	return &GetActionTemplateByIdResponse{
		Template: *out,
	}, nil
}

const DeleteActionTemplateRequestSubject = "request.cueing.actiontemplate.delete"

type DeleteActionTemplateRequest struct {
	TemplateId string `msgpack:"templateId" json:"templateId" validate:"required"`
}

type DeleteActionTemplateResponse struct{}

func (c *CueingApi) DeleteActionTemplate(sub string, request *DeleteActionTemplateRequest) (*DeleteActionTemplateResponse, error) {
	err := c.model.DeleteActionTemplateById(request.TemplateId)
	if err != nil {
		return nil, err
	}

	return &DeleteActionTemplateResponse{}, nil
}
