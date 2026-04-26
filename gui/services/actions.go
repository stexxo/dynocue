// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package services

import (
	"errors"

	"github.com/stexxo/dynocue/client"
	"github.com/stexxo/dynocue/components/cues"
	"github.com/stexxo/dynocue/components/cues/types"
	"github.com/stexxo/dynocue/core/logging"
	"github.com/wailsapp/wails/v3/pkg/application"
)

type ActionsService struct {
	clientManager *client.Manager
	app           *application.App
	logger        logging.Logger
}

func NewActionsService(manager *client.Manager, app *application.App, logger logging.Logger) *ActionsService {
	out := &ActionsService{
		app:           app,
		logger:        logger,
		clientManager: manager,
	}
	out.clientManager.OnNewClient(out.onNewClient)
	return out
}

func (s *ActionsService) onNewClient(cl *client.Client) error {
	return errors.Join(
		cl.OnActionCreated(func(sub string, event *cues.ActionCreatedEvent) { s.app.Event.Emit(sub, event) }),
		cl.OnActionUpdated(func(sub string, event *cues.ActionUpdatedEvent) { s.app.Event.Emit(sub, event) }),
		cl.OnActionRenumbered(func(sub string, event *cues.ActionRenumberedEvent) { s.app.Event.Emit(sub, event) }),
		cl.OnActionDeleted(func(sub string, event *cues.ActionDeletedEvent) { s.app.Event.Emit(sub, event) }),
	)
}

func (s *ActionsService) CreateAction(cueListId string, cueId string, templateId string, number float64) (*types.Action, bool) {
	var out *types.Action
	err := s.clientManager.WithClient(func(c *client.Client) error {
		action, err := c.CreateAction(cueListId, cueId, templateId, number)
		if err != nil {
			return err
		}
		out = action
		return nil
	})

	if err != nil {
		s.logger.Error("failed to create action", "err", err)
		return nil, false
	}

	return out, true
}

func (s *ActionsService) EnumerateActions(cueListId string, cueId string) ([]types.Action, bool) {
	var out []types.Action
	err := s.clientManager.WithClient(func(c *client.Client) error {
		actions, err := c.EnumerateActions(cueListId, cueId)
		if err != nil {
			return err
		}
		out = actions
		return nil
	})

	if err != nil {
		s.logger.Error("failed to enumerate actions", "err", err)
		return nil, false
	}

	return out, true
}

func (s *ActionsService) GetActionByNumber(cueListId string, cueId string, number float64) (*types.Action, bool) {
	var out *types.Action
	err := s.clientManager.WithClient(func(c *client.Client) error {
		action, err := c.GetActionByNumber(cueListId, cueId, number)
		if err != nil {
			return err
		}
		out = action
		return nil
	})

	if err != nil {
		s.logger.Error("failed to get action by number", "err", err)
		return nil, false
	}

	return out, true
}

func (s *ActionsService) GetActionById(cueListId string, cueId string, actionId string) (*types.Action, bool) {
	var out *types.Action
	err := s.clientManager.WithClient(func(c *client.Client) error {
		action, err := c.GetActionById(cueListId, cueId, actionId)
		if err != nil {
			return err
		}
		out = action
		return nil
	})

	if err != nil {
		s.logger.Error("failed to get action by id", "err", err)
		return nil, false
	}

	return out, true
}

func (s *ActionsService) UpdateAction(cueListId string, cueId string, actionId string, field string, value any) (*types.Action, bool) {
	var out *types.Action
	err := s.clientManager.WithClient(func(c *client.Client) error {
		action, err := c.UpdateAction(cueListId, cueId, actionId, field, value)
		if err != nil {
			return err
		}
		out = action
		return nil
	})

	if err != nil {
		s.logger.Error("failed to update action", "err", err)
		return nil, false
	}

	return out, true
}

func (s *ActionsService) UpdateActionField(cueListId string, cueId string, actionId string, fieldName string, value any) (*types.Action, bool) {
	var out *types.Action
	err := s.clientManager.WithClient(func(c *client.Client) error {
		action, err := c.UpdateActionField(cueListId, cueId, actionId, fieldName, value)
		if err != nil {
			return err
		}
		out = action
		return nil
	})

	if err != nil {
		s.logger.Error("failed to update action field", "err", err)
		return nil, false
	}

	return out, true
}

func (s *ActionsService) RenumberAction(cueListId string, cueId string, actionId string, newNumber float64) bool {
	err := s.clientManager.WithClient(func(c *client.Client) error {
		return c.RenumberAction(cueListId, cueId, actionId, newNumber)
	})

	if err != nil {
		s.logger.Error("failed to renumber action", "err", err)
		return false
	}

	return true
}

func (s *ActionsService) DeleteAction(cueListId string, cueId string, actionId string) bool {
	err := s.clientManager.WithClient(func(c *client.Client) error {
		return c.DeleteAction(cueListId, cueId, actionId)
	})

	if err != nil {
		s.logger.Error("failed to delete action", "err", err)
		return false
	}

	return true
}
