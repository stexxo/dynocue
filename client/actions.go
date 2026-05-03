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

var ErrActionExists = fmt.Errorf("action with provided number already exists")
var ErrActionNotFound = fmt.Errorf("action not found")

func (c *Client) CreateAction(cueId string, templateId string) (*types.Action, error) {
	resp, err := messaging.Request[cues.CreateActionResponse](c.messenger, cues.CreateActionRequestSubject, &cues.CreateActionRequest{
		CueId:      cueId,
		TemplateId: templateId,
	})
	if err != nil {
		return nil, err
	}
	if resp.Success {
		return &resp.Response.Action, nil
	}

	if resp.Error == cues.ActionNumberExists {
		return nil, ErrActionExists
	}

	return nil, fmt.Errorf("failed to create action: %s", resp.Error)
}

func (c *Client) EnumerateActions(cueId string) ([]types.Action, error) {
	resp, err := messaging.Request[cues.EnumerateActionsResponse](c.messenger, cues.EnumerateActionsRequestSubject, &cues.EnumerateActionsRequest{
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
	resp, err := messaging.Request[cues.GetActionByIdResponse](c.messenger, cues.GetActionByIdRequestSubject, &cues.GetActionByIdRequest{
		ActionId: actionId,
	})
	if err != nil {
		return nil, err
	}
	if resp.Success {
		return &resp.Response.Action, nil
	}

	if resp.Error == cues.ActionNotFound {
		return nil, ErrActionNotFound
	}

	return nil, fmt.Errorf("failed to get action: %s", resp.Error)
}

func (c *Client) UpdateAction(actionId string, field string, value any) error {
	resp, err := messaging.Request[cues.UpdateActionResponse](c.messenger, cues.UpdateActionRequestSubject, &cues.UpdateActionRequest{
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

	if resp.Error == cues.ActionNotFound {
		return ErrActionNotFound
	}

	return fmt.Errorf("failed to update action: %s", resp.Error)
}

func (c *Client) UpdateActionField(actionId string, fieldName string, value any) error {
	resp, err := messaging.Request[cues.UpdateActionFieldResponse](c.messenger, cues.UpdateActionFieldRequestSubject, &cues.UpdateActionFieldRequest{
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

	if resp.Error == cues.ActionNotFound {
		return ErrActionNotFound
	}

	return fmt.Errorf("failed to update action field: %s", resp.Error)
}

func (c *Client) DeleteAction(actionId string) error {
	resp, err := messaging.Request[cues.DeleteActionResponse](c.messenger, cues.DeleteActionRequestSubject, &cues.DeleteActionRequest{
		ActionId: actionId,
	})
	if err != nil {
		return err
	}
	if resp.Success {
		return nil
	}

	if resp.Error == cues.ActionNotFound {
		return ErrActionNotFound
	}

	return fmt.Errorf("failed to delete action: %s", resp.Error)
}

func (c *Client) OnActionCreated(handler EventCallback[cues.ActionCreatedEvent]) error {
	return messaging.Subscribe[cues.ActionCreatedEvent](c.messenger, true, cues.ActionCreatedEventSubject, func(sub string, msg *cues.ActionCreatedEvent) {
		handler(sub, msg)
	})
}

func (c *Client) OnActionUpdated(handler EventCallback[cues.ActionUpdatedEvent]) error {
	return messaging.Subscribe[cues.ActionUpdatedEvent](c.messenger, true, cues.ActionUpdatedEventSubject, func(sub string, msg *cues.ActionUpdatedEvent) {
		handler(sub, msg)
	})
}

func (c *Client) OnActionDeleted(handler EventCallback[cues.ActionDeletedEvent]) error {
	return messaging.Subscribe[cues.ActionDeletedEvent](c.messenger, true, cues.ActionDeletedEventSubject, func(sub string, msg *cues.ActionDeletedEvent) {
		handler(sub, msg)
	})
}
