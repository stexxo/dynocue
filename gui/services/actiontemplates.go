// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package services

import (
	"github.com/stexxo/dynocue/client"
	"github.com/stexxo/dynocue/components/cues"
	"github.com/stexxo/dynocue/components/cues/types"
	"github.com/stexxo/dynocue/core/logging"
	"github.com/wailsapp/wails/v3/pkg/application"
)

type ActionTemplatesService struct {
	clientManager *client.Manager
	app           *application.App
	logger        logging.Logger
}

func NewActionTemplatesService(manager *client.Manager, app *application.App, logger logging.Logger) *ActionTemplatesService {
	out := &ActionTemplatesService{
		app:           app,
		logger:        logger,
		clientManager: manager,
	}
	out.clientManager.OnNewClient(out.onNewClient)
	return out
}

func (s *ActionTemplatesService) onNewClient(cl *client.Client) error {
	return cl.OnActionTemplateRegistered(func(subject string, event *cues.RegisterActionTemplateEvent) {
		s.app.Event.Emit(subject, event)
	})
}

func (s *ActionTemplatesService) EnumerateActionTemplates() ([]types.ActionTemplate, bool) {
	var out []types.ActionTemplate
	err := s.clientManager.WithClient(func(c *client.Client) error {
		templates, err := c.EnumerateActionTemplates()
		if err != nil {
			return err
		}
		out = templates
		return nil
	})

	if err != nil {
		s.logger.Error("failed to enumerate action templates", "err", err)
		return nil, false
	}

	return out, true
}

func (s *ActionTemplatesService) GetActionTemplate(id string) (*types.ActionTemplate, bool) {
	var out *types.ActionTemplate
	err := s.clientManager.WithClient(func(c *client.Client) error {
		template, err := c.GetActionTemplate(id)
		if err != nil {
			return err
		}
		out = template
		return nil
	})

	if err != nil {
		s.logger.Error("failed to get action template", "err", err, "id", id)
		return nil, false
	}

	return out, true
}
