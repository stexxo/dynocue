// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package client

import (
	"fmt"

	"github.com/stexxo/dynocue/components/cues"
	"github.com/stexxo/dynocue/components/cues/types"
	"github.com/stexxo/dynocue/core/messaging"
)

var ErrActionTemplateNotFound = fmt.Errorf("action template not found")

func (c *Client) EnumerateActionTemplates() ([]types.ActionTemplate, error) {
	resp, err := messaging.Request[cues.EnumerateActionTemplatesResponse](c.messenger, cues.EnumerateActionTemplatesRequestSubject, &cues.EnumerateActionTemplatesRequest{})
	if err != nil {
		return nil, err
	}
	if resp.Success {
		return resp.Response.ActionTemplates, nil
	}

	return nil, fmt.Errorf("failed to enumerate action templates: %s", resp.Error)
}

func (c *Client) GetActionTemplate(id string) (*types.ActionTemplate, error) {
	resp, err := messaging.Request[cues.GetActionTemplateResponse](c.messenger, cues.GetActionTemplateRequestSubject, &cues.GetActionTemplateRequest{
		Id: id,
	})
	if err != nil {
		return nil, err
	}
	if resp.Success {
		return resp.Response.Template, nil
	}

	if resp.Error == cues.ActionTemplateNotFound {
		return nil, ErrActionTemplateNotFound
	}

	return nil, fmt.Errorf("failed to get action template: %s", resp.Error)
}

func (c *Client) RegisterActionTemplate(id, subsystemName, name, subject string, fields []types.ActionTemplateField) error {
	resp, err := messaging.Request[cues.RegisterActionTemplateResponse](c.messenger, cues.RegisterActionTemplateRequestSubject, &cues.RegisterActionTemplateRequest{
		Id:            id,
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

func (c *Client) OnActionTemplateRegistered(handler EventCallback[cues.RegisterActionTemplateEvent]) error {
	return messaging.Subscribe[cues.RegisterActionTemplateEvent](c.messenger, true, cues.RegisterActionTemplateEventSubject, func(sub string, msg *cues.RegisterActionTemplateEvent) {
		handler(sub, msg)
	})
}
