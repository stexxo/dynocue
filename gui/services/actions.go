// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package services

import (
	"errors"

	"github.com/stexxo/dynocue/client"
	"github.com/stexxo/dynocue/components/cues/api"
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
		cl.OnActionCreated(func(sub string, event *api.ActionChangeEvent) { s.app.Event.Emit(sub, event) }),
		cl.OnActionUpdated(func(sub string, event *api.ActionChangeEvent) { s.app.Event.Emit(sub, event) }),
		cl.OnActionDeleted(func(sub string, event *api.ActionChangeEvent) { s.app.Event.Emit(sub, event) }),
	)
}

func (s *ActionsService) CreateAction(cueId string, templateId string, actionNumber uint) bool {
	err := s.clientManager.WithClient(func(c *client.Client) error {
		_, _, err := c.CreateAction(cueId, templateId, actionNumber)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		s.logger.Error("failed to create action", "err", err)
		return false
	}

	return true
}

func (s *ActionsService) EnumerateActions(cueId string) ([]types.Action, bool) {
	var out []types.Action
	err := s.clientManager.WithClient(func(c *client.Client) error {
		actions, err := c.EnumerateActions(cueId)
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

func (s *ActionsService) GetActionById(actionId string) (*types.Action, bool) {
	var out *types.Action
	err := s.clientManager.WithClient(func(c *client.Client) error {
		action, err := c.GetActionById(actionId)
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

func (s *ActionsService) UpdateAction(actionId string, field string, value any) bool {
	err := s.clientManager.WithClient(func(c *client.Client) error {
		return c.UpdateAction(actionId, field, value)
	})

	if err != nil {
		s.logger.Error("failed to update action", "err", err)
		return false
	}

	return true
}

func (s *ActionsService) UpdateActionField(actionId string, fieldName string, value any) bool {
	err := s.clientManager.WithClient(func(c *client.Client) error {
		return c.UpdateActionField(actionId, fieldName, value)
	})

	if err != nil {
		s.logger.Error("failed to update action field", "err", err)
		return false
	}

	return true
}

func (s *ActionsService) DeleteAction(actionId string) bool {
	err := s.clientManager.WithClient(func(c *client.Client) error {
		return c.DeleteAction(actionId)
	})

	if err != nil {
		s.logger.Error("failed to delete action", "err", err)
		return false
	}

	return true
}
