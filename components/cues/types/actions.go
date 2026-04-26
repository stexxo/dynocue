package types

import (
	"sync"
	"time"
)

type Action struct {
	Subject      string    `msgpack:"subject" json:"subject"`
	TemplateType string    `msgpack:"templateType" json:"templateType"`
	Delay        time.Time `msgpack:"delay" json:"delay"`
	Follow       time.Time `msgpack:"follow" json:"follow"`
}

type ActionFields struct {
	FieldName string      `msgpack:"fieldName" json:"fieldName"`
	Value     interface{} `msgpack:"value" json:"value"`
}

type ActionTemplate struct {
	Id           string                `msgpack:"id" json:"id"`
	TemplateName string                `msgpack:"templateName" json:"templateName"`
	Subject      string                `msgpack:"subject" json:"subject"`
	Fields       []ActionTemplateField `msgpack:"fields" json:"fields"`
}

type ActionTemplateField struct {
	FieldName  string `msgpack:"fieldName" json:"fieldName"`
	FieldLabel string `msgpack:"fieldLabel" json:"fieldLabel"`
	DataType   string `msgpack:"dataType" json:"dataType"` // string, float, int, bool, []string, []float, []int
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
