// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package types

type ActionTemplate struct {
	TemplateId    string                `msgpack:"templateId" json:"templateId"`
	TemplateName  string                `msgpack:"templateName" json:"templateName"`
	SubsystemName string                `msgpack:"subsystemName" json:"subsystemName"`
	Subject       string                `msgpack:"subject" json:"subject"`
	Fields        []ActionTemplateField `msgpack:"fields" json:"fields"`
}

func (a *ActionTemplate) NewAction(cueListId string, cueId string) *Action {
	action := NewAction(cueListId, cueId)
	action.TemplateId = a.TemplateId
	action.Subject = a.Subject

	for _, f := range a.Fields {
		action.Fields = append(action.Fields, ActionFields{FieldName: f.FieldName, FieldLabel: f.FieldLabel, DataType: f.DataType, Value: f.DefaultValue})
	}

	return action
}

type ActionTemplateField struct {
	FieldName    string      `msgpack:"fieldName" json:"fieldName"`
	FieldLabel   string      `msgpack:"fieldLabel" json:"fieldLabel"`
	DataType     string      `msgpack:"dataType" json:"dataType"` // string, float, int, bool, time
	DefaultValue interface{} `msgpack:"defaultValue" json:"defaultValue"`
}
