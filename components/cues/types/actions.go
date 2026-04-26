// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package types

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

type Action struct {
	Id         string         `msgpack:"id" json:"id"`
	Label      string         `msgpack:"label" json:"label"`
	TemplateId string         `msgpack:"templateId" json:"templateId"`
	Delay      time.Duration  `msgpack:"delay" json:"delay"`
	Fields     []ActionFields `msgpack:"fields" json:"fields"`
}

func NewActionByTemplate(actionTemplate *ActionTemplate) *Action {
	action := NewAction()
	action.TemplateId = actionTemplate.Id

	for _, f := range actionTemplate.Fields {
		action.Fields = append(action.Fields, ActionFields{FieldName: f.FieldName, FieldLabel: f.FieldLabel, DataType: f.DataType, Value: f.DefaultValue})
	}

	return action
}

func NewAction() *Action {
	return &Action{
		Id: uuid.NewString(),
	}
}

type ActionFields struct {
	FieldName  string      `msgpack:"fieldName" json:"fieldName"`
	FieldLabel string      `msgpack:"fieldLabel" json:"fieldLabel"`
	DataType   string      `msgpack:"dataType" json:"dataType"`
	Value      interface{} `msgpack:"value" json:"value"`
}

type ActionTemplate struct {
	Id            string                `msgpack:"id" json:"id"`
	TemplateName  string                `msgpack:"templateName" json:"templateName"`
	SubsystemName string                `msgpack:"subsystemName" json:"subsystemName"`
	Subject       string                `msgpack:"subject" json:"subject"`
	Fields        []ActionTemplateField `msgpack:"fields" json:"fields"`
}

type ActionTemplateField struct {
	FieldName    string      `msgpack:"fieldName" json:"fieldName"`
	FieldLabel   string      `msgpack:"fieldLabel" json:"fieldLabel"`
	DataType     string      `msgpack:"dataType" json:"dataType"` // string, float, int, bool, time
	DefaultValue interface{} `msgpack:"defaultValue" json:"defaultValue"`
}

type ActionTemplatesModel struct {
	mu        sync.RWMutex
	templates []ActionTemplate
}

func NewActionTemplatesModel() *ActionTemplatesModel {
	return &ActionTemplatesModel{}
}

func (p *ActionTemplatesModel) AddTemplate(template ActionTemplate) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.templates = append(p.templates, template)
}

func (p *ActionTemplatesModel) GetTemplates() []ActionTemplate {
	p.mu.RLock()
	defer p.mu.RUnlock()
	out := make([]ActionTemplate, len(p.templates))
	copy(out, p.templates)
	return out
}

func (p *ActionTemplatesModel) GetTemplateById(id string) *ActionTemplate {
	p.mu.RLock()
	defer p.mu.RUnlock()
	for _, template := range p.templates {
		if template.Id == id {
			return &template
		}
	}
	return nil
}
