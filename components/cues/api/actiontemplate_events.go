// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package api

import (
	"github.com/stexxo/dynocue/components/cues/model"
	"github.com/stexxo/dynocue/util"
)

func (c *CueingApi) registerActionTemplateEvents() {
	c.model.RegisterEventHandler(model.ResourceActionTemplate, model.OperationCreated, eventHandler[ActionTemplateChangeEvent](c.messenger, c.logger, c.ActionTemplateChanged))
	c.model.RegisterEventHandler(model.ResourceActionTemplate, model.OperationDeleted, eventHandler[ActionTemplateChangeEvent](c.messenger, c.logger, c.ActionTemplateChanged))
}

const (
	ActionTemplateCreatedEventSubject = "event.cueing.actions.templates.created"
	DeleteActionTemplateEventSubject  = "event.cueing.actions.templates.deleted"
)

type ActionTemplateChangeEvent struct {
	TemplateId string `msgpack:"templateId" json:"templateId"`
}

func (c *CueingApi) ActionTemplateChanged(ev util.Event) (string, *ActionTemplateChangeEvent) {
	var sub string
	switch ev.Operation {
	case model.OperationDeleted:
		sub = DeleteActionTemplateEventSubject
	case model.OperationCreated:
		sub = ActionTemplateCreatedEventSubject
	}
	return sub, &ActionTemplateChangeEvent{
		TemplateId: ev.EventData[model.MetadataActionTemplateId],
	}
}
