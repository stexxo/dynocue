// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package services

import (
	"errors"

	"github.com/stexxo/dynocue/client"
	"github.com/stexxo/dynocue/components/cues/api"
	"github.com/stexxo/dynocue/core/logging"
	"github.com/wailsapp/wails/v3/pkg/application"
)

type ExecutionService struct {
	clientManager *client.Manager
	app           *application.App
	logger        logging.Logger
}

func NewExecutionService(manager *client.Manager, app *application.App, logger logging.Logger) *ExecutionService {
	out := &ExecutionService{
		app:           app,
		logger:        logger,
		clientManager: manager,
	}
	out.clientManager.OnNewClient(out.onNewClient)
	return out
}

func (c *ExecutionService) onNewClient(cl *client.Client) error {
	return errors.Join(
		cl.OnExecutionStarted(func(s string, e *api.ExecutionChangeEvent) { c.app.Event.Emit(s, e) }),
		cl.OnExecutionFinished(func(s string, e *api.ExecutionChangeEvent) { c.app.Event.Emit(s, e) }),
		cl.OnExecutionUnselected(func(s string, e *api.ExecutionChangeEvent) { c.app.Event.Emit(s, e) }),
		cl.OnExecutionDeleted(func(s string, e *api.ExecutionChangeEvent) { c.app.Event.Emit(s, e) }),
	)
}

func (c *ExecutionService) GoToCue(cueId string) bool {
	err := c.clientManager.WithClient(func(c *client.Client) error {
		return c.GoToCue(cueId)
	})

	if err != nil {
		c.logger.Error("failed to go to cue", "err", err)
		return false
	}

	return true
}

func (c *ExecutionService) GoToNextCue(cueListId string) bool {
	err := c.clientManager.WithClient(func(c *client.Client) error {
		return c.GoToNextCue(cueListId)
	})

	if err != nil {
		c.logger.Error("failed to go to next cue", "err", err)
		return false
	}

	return true
}
