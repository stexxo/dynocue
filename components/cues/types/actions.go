package types

import "time"

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
	TemplateName string                `msgpack:"templateName" json:"templateName"`
	Subject      string                `msgpack:"subject" json:"subject"`
	Fields       []ActionTemplateField `msgpack:"fields" json:"fields"`
}

type ActionTemplateField struct {
	FieldName string `msgpack:"fieldName" json:"fieldName"`
	DataType  string `msgpack:"dataType" json:"dataType"` // string, float, int, bool, []string, []float, []int
}
