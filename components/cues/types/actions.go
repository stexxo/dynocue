package types

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

type Action struct {
	Id         string         `msgpack:"id" json:"id"`
	Number     float64        `msgpack:"number" json:"number"`
	Subject    string         `msgpack:"subject" json:"subject"`
	TemplateId string         `msgpack:"templateId" json:"templateId"`
	Delay      time.Time      `msgpack:"delay" json:"delay"`
	Follow     time.Time      `msgpack:"follow" json:"follow"`
	Fields     []ActionFields `msgpack:"fields" json:"fields"`
}

func NewActionByTemplate(actionTemplate *ActionTemplate) *Action {
	action := NewAction()
	action.Subject = actionTemplate.Subject

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

func (action *Action) Num() float64 {
	return action.Number
}

func (action *Action) SetNum(number float64) {
	action.Number = number
}

type ActionFields struct {
	FieldName  string      `msgpack:"fieldName" json:"fieldName"`
	FieldLabel string      `msgpack:"fieldType" json:"fieldType"`
	DataType   string      `msgpack:"dataType" json:"dataType"`
	Value      interface{} `msgpack:"value" json:"value"`
}

type ActionTemplate struct {
	Id           string                `msgpack:"id" json:"id"`
	TemplateName string                `msgpack:"templateName" json:"templateName"`
	Subject      string                `msgpack:"subject" json:"subject"`
	Fields       []ActionTemplateField `msgpack:"fields" json:"fields"`
}

type ActionTemplateField struct {
	FieldName    string      `msgpack:"fieldName" json:"fieldName"`
	FieldLabel   string      `msgpack:"fieldLabel" json:"fieldLabel"`
	DataType     string      `msgpack:"dataType" json:"dataType"` // string, float, int, bool
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
