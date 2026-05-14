// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package client

import (
	"fmt"

	"github.com/stexxo/dynocue/components/cues/api"
	"github.com/stexxo/dynocue/components/cues/types"
	"github.com/stexxo/dynocue/core/messaging"
)

var ErrActionExists = fmt.Errorf("action with provided number already exists")
var ErrActionNotFound = fmt.Errorf("action not found")

func (c *Client) CreateAction(cueId string, templateId string, actionNumber uint) (string, uint, error) {
	resp, err := messaging.Request[api.CreateActionResponse](c.messenger, api.CreateActionRequestSubject, &api.CreateActionRequest{
		CueId:      cueId,
		TemplateId: templateId,
		Number:     actionNumber,
	})
	if err != nil {
		return "", 0, err
	}
	if resp.Success {
		return resp.Response.ActionId, resp.Response.Number, nil
	}

	if resp.Error == api.ActionNumberExists {
		return "", 0, ErrActionExists
	}

	return "", 0, fmt.Errorf("failed to create action: %s", resp.Error)
}

func (c *Client) EnumerateActions(cueId string) ([]types.Action, error) {
	resp, err := messaging.Request[api.EnumerateActionsResponse](c.messenger, api.EnumerateActionsRequestSubject, &api.EnumerateActionsRequest{
		CueId: cueId,
	})
	if err != nil {
		return nil, err
	}
	if resp.Success {
		return resp.Response.Actions, nil
	}

	return nil, fmt.Errorf("failed to enumerate actions: %s", resp.Error)
}

func (c *Client) GetActionById(actionId string) (*types.Action, error) {
	resp, err := messaging.Request[api.GetActionByIdResponse](c.messenger, api.GetActionByIdRequestSubject, &api.GetActionByIdRequest{
		ActionId: actionId,
	})
	if err != nil {
		return nil, err
	}
	if resp.Success {
		return &resp.Response.Action, nil
	}

	if resp.Error == api.ActionNotFound {
		return nil, ErrActionNotFound
	}

	return nil, fmt.Errorf("failed to get action: %s", resp.Error)
}

func (c *Client) UpdateAction(actionId string, field string, value any) error {
	resp, err := messaging.Request[api.UpdateActionResponse](c.messenger, api.UpdateActionRequestSubject, &api.UpdateActionRequest{
		ActionId: actionId,
		Field:    field,
		Value:    value,
	})
	if err != nil {
		return err
	}
	if resp.Success {
		return nil
	}

	if resp.Error == api.ActionNotFound {
		return ErrActionNotFound
	}

	return fmt.Errorf("failed to update action: %s", resp.Error)
}

func (c *Client) UpdateActionField(actionId string, fieldName string, value any) error {
	resp, err := messaging.Request[api.UpdateActionFieldResponse](c.messenger, api.UpdateActionFieldRequestSubject, &api.UpdateActionFieldRequest{
		ActionId:  actionId,
		FieldName: fieldName,
		Value:     value,
	})
	if err != nil {
		return err
	}
	if resp.Success {
		return nil
	}

	if resp.Error == api.ActionNotFound {
		return ErrActionNotFound
	}

	return fmt.Errorf("failed to update action field: %s", resp.Error)
}

func (c *Client) DeleteAction(actionId string) error {
	resp, err := messaging.Request[api.DeleteActionResponse](c.messenger, api.DeleteActionRequestSubject, &api.DeleteActionRequest{
		ActionId: actionId,
	})
	if err != nil {
		return err
	}
	if resp.Success {
		return nil
	}

	if resp.Error == api.ActionNotFound {
		return ErrActionNotFound
	}

	return fmt.Errorf("failed to delete action: %s", resp.Error)
}

func (c *Client) OnActionCreated(handler EventCallback[api.ActionChangeEvent]) error {
	return messaging.Subscribe[api.ActionChangeEvent](c.messenger, true, api.ActionCreatedEventSubject, func(sub string, msg *api.ActionChangeEvent) {
		handler(sub, msg)
	})
}

func (c *Client) OnActionUpdated(handler EventCallback[api.ActionChangeEvent]) error {
	return messaging.Subscribe[api.ActionChangeEvent](c.messenger, true, api.ActionAttributesUpdatedEventSubject, func(sub string, msg *api.ActionChangeEvent) {
		handler(sub, msg)
	})
}

func (c *Client) OnActionDeleted(handler EventCallback[api.ActionChangeEvent]) error {
	return messaging.Subscribe[api.ActionChangeEvent](c.messenger, true, api.ActionDeletedEventSubject, func(sub string, msg *api.ActionChangeEvent) {
		handler(sub, msg)
	})
}
