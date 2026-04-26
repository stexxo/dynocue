package cues

import (
	"github.com/google/uuid"
	"github.com/stexxo/dynocue/components/cues/types"
	"github.com/stexxo/dynocue/core/messaging"
)

const ActionTemplateNotFound = "Action template not found."

const RegisterActionTemplateRequestSubject = "request.cueing.actions.templates.register"
const RegisterActionTemplateEventSubject = "event.cueing.actions.templates.registered"

type RegisterActionTemplateRequest struct {
	Id      string                      `msgpack:"id" json:"id"`
	Name    string                      `msgpack:"name" json:"name"`
	Subject string                      `msgpack:"subject" json:"subject"`
	Fields  []types.ActionTemplateField `msgpack:"fields" json:"fields"`
}

type RegisterActionTemplateResponse struct{}

type RegisterActionTemplateEvent struct {
	Id   string `msgpack:"id" json:"id"`
	Name string `msgpack:"name" json:"name"`
}

func (p *Cueing) RegisterActionType(sub string, req *RegisterActionTemplateRequest) (*RegisterActionTemplateResponse, error) {
	p.actionTemplates.AddTemplate(types.ActionTemplate{Id: req.Id, TemplateName: req.Name, Subject: req.Subject, Fields: req.Fields})
	err := messaging.Publish(p.Messenger(), RegisterActionTemplateEventSubject, &RegisterActionTemplateEvent{Id: uuid.NewString(), Name: req.Name})
	if err != nil {
		return nil, err
	}
	return &RegisterActionTemplateResponse{}, nil
}

const EnumerateActionTemplatesRequestSubject = "request.cueing.actions.templates.enumerate"

type EnumerateActionTemplatesRequest struct{}
type EnumerateActionTemplatesResponse struct {
	ActionTemplates []types.ActionTemplate `msgpack:"actionTemplates" json:"actionTemplates"`
}

func (p *Cueing) EnumerateActionTemplates(sub string, req *EnumerateActionTemplatesRequest) (*EnumerateActionTemplatesResponse, error) {
	out := p.actionTemplates.GetTemplates()
	return &EnumerateActionTemplatesResponse{out}, nil
}

const GetActionTemplateRequestSubject = "request.cueing.actions.templates.get"

type GetActionTemplateRequest struct {
	Id string
}
type GetActionTemplateResponse struct {
	Template *types.ActionTemplate `msgpack:"template" json:"template"`
}

func (p *Cueing) GetActionTemplate(sub string, req *GetActionTemplateRequest) (*GetActionTemplateResponse, error) {
	template := p.actionTemplates.GetTemplateById(req.Id)
	if template == nil {
		return nil, &messaging.FriendlyError{FriendlyErr: ActionTemplateNotFound}
	}

	return &GetActionTemplateResponse{Template: template}, nil
}
