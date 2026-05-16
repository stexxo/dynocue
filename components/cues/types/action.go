// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package types

import (
	"time"

	"github.com/google/uuid"
)

type Action struct {
	CueListId     string         `msgpack:"cueListId" json:"cueListId"`
	CueId         string         `msgpack:"cueId" json:"cueId"`
	ActionId      string         `msgpack:"actionId" json:"actionId"`
	Number        uint           `msgpack:"number" json:"number"`
	Subject       string         `msgpack:"subject" json:"subject"`
	Label         string         `msgpack:"label" json:"label"`
	TemplateId    string         `msgpack:"templateId" json:"templateId"`
	WaitForFinish bool           `msgpack:"waitForFinish" json:"waitForFinish"`
	Delay         time.Duration  `msgpack:"delay" json:"delay"`
	Fields        []ActionFields `msgpack:"fields" json:"fields"`
}

func NewAction(cueId string) *Action {
	return &Action{
		ActionId: uuid.NewString(),
		CueId:    cueId,
	}
}

type ActionFields struct {
	FieldName  string      `msgpack:"fieldName" json:"fieldName"`
	FieldLabel string      `msgpack:"fieldLabel" json:"fieldLabel"`
	DataType   string      `msgpack:"dataType" json:"dataType"`
	Value      interface{} `msgpack:"value" json:"value"`
}
