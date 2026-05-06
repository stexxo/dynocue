// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cues

import (
	"github.com/hashicorp/go-memdb"
	"github.com/stexxo/dynocue/components/cues/types"
	"github.com/stexxo/dynocue/core/messaging"
	"github.com/stexxo/dynocue/db"
)

const ActionTemplateNotFound = "Action template not found."

const RegisterActionTemplateRequestSubject = "request.cueing.actions.templates.register"
const RegisterActionTemplateEventSubject = "event.cueing.actions.templates.registered"

type RegisterActionTemplateRequest struct {
	TemplateId    string                      `msgpack:"templateId" json:"templateId"`
	SubsystemName string                      `msgpack:"subsystemName" json:"subsystemName"`
	Name          string                      `msgpack:"name" json:"name"`
	Subject       string                      `msgpack:"subject" json:"subject"`
	Fields        []types.ActionTemplateField `msgpack:"fields" json:"fields"`
}

type RegisterActionTemplateResponse struct{}

type RegisterActionTemplateEvent struct {
	TemplateId string `msgpack:"templateId" json:"templateId"`
	Name       string `msgpack:"name" json:"name"`
}

func (p *Cueing) RegisterActionType(sub string, req *RegisterActionTemplateRequest) (*RegisterActionTemplateResponse, error) {
	err := db.WithWrite(p.runtimeDb, func(txn *memdb.Txn) error {
		return txn.Insert(TableActionTemplates, &types.ActionTemplate{
			TemplateId:    req.TemplateId,
			TemplateName:  req.Name,
			Subject:       req.Subject,
			Fields:        req.Fields,
			SubsystemName: req.SubsystemName,
		})
	})
	if err != nil {
		return nil, err
	}

	err = messaging.Publish(p.Messenger(), RegisterActionTemplateEventSubject, &RegisterActionTemplateEvent{TemplateId: req.TemplateId, Name: req.Name})
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
	out, err := db.GetAllDb[types.ActionTemplate](p.runtimeDb, TableActionTemplates, IndexId)
	if err != nil {
		return nil, err
	}
	return &EnumerateActionTemplatesResponse{ActionTemplates: out}, nil
}

const GetActionTemplateRequestSubject = "request.cueing.actions.templates.get"

type GetActionTemplateRequest struct {
	TemplateId string `msgpack:"templateId" json:"templateId"`
}
type GetActionTemplateResponse struct {
	Template *types.ActionTemplate `msgpack:"template" json:"template"`
}

func (p *Cueing) GetActionTemplate(sub string, req *GetActionTemplateRequest) (*GetActionTemplateResponse, error) {
	template, err := db.GetFirstDb[types.ActionTemplate](p.runtimeDb, TableActionTemplates, IndexId, req.TemplateId)
	if err != nil {
		return nil, &messaging.FriendlyError{FriendlyErr: ActionTemplateNotFound}
	}

	return &GetActionTemplateResponse{Template: template}, nil
}
