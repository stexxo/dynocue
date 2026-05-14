// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package client

import (
	"fmt"

	"github.com/stexxo/dynocue/components/cues"
	"github.com/stexxo/dynocue/components/cues/api"
	"github.com/stexxo/dynocue/components/cues/types"
	"github.com/stexxo/dynocue/core/messaging"
)

var ErrActionTemplateNotFound = fmt.Errorf("action template not found")

func (c *Client) EnumerateActionTemplates() ([]types.ActionTemplate, error) {
	resp, err := messaging.Request[api.EnumerateActionTemplatesResponse](c.messenger, api.EnumerateActionTemplatesRequestSubject, &api.EnumerateActionTemplatesRequest{})
	if err != nil {
		return nil, err
	}
	if resp.Success {
		return resp.Response.Templates, nil
	}

	return nil, fmt.Errorf("failed to enumerate action templates: %s", resp.Error)
}

func (c *Client) GetActionTemplate(id string) (*types.ActionTemplate, error) {
	resp, err := messaging.Request[api.GetActionTemplateByIdResponse](c.messenger, api.GetActionTemplateByIdRequestSubject, &api.GetActionTemplateByIdRequest{
		TemplateId: id,
	})
	if err != nil {
		return nil, err
	}
	if resp.Success {
		return &resp.Response.Template, nil
	}

	if resp.Error == api.ActionTemplateNotFound {
		return nil, ErrActionTemplateNotFound
	}

	return nil, fmt.Errorf("failed to get action template: %s", resp.Error)
}

func (c *Client) RegisterActionTemplate(id, subsystemName, name, subject string, fields []types.ActionTemplateField) error {
	resp, err := messaging.Request[api.RegisterActionTemplateResponse](c.messenger, api.RegisterActionTemplateRequestSubject, &api.RegisterActionTemplateRequest{
		TemplateId:    id,
		SubsystemName: subsystemName,
		Name:          name,
		Subject:       subject,
		Fields:        fields,
	})
	if err != nil {
		return err
	}
	if resp.Success {
		return nil
	}

	return fmt.Errorf("failed to register action template: %s", resp.Error)
}

func (c *Client) OnActionTemplateRegistered(handler EventCallback[api.RegisterActionTemplateEvent]) error {
	return messaging.Subscribe[api.RegisterActionTemplateEvent](c.messenger, true, api.RegisterActionTemplateEventSubject, func(sub string, msg *api.RegisterActionTemplateEvent) {
		handler(sub, msg)
	})
}
