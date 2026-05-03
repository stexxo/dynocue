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

type CueListsService struct {
	clientManager *client.Manager
	app           *application.App
	logger        logging.Logger
}

func NewCueListsService(manager *client.Manager, app *application.App, logger logging.Logger) *CueListsService {
	out := &CueListsService{
		app:           app,
		logger:        logger,
		clientManager: manager,
	}
	out.clientManager.OnNewClient(out.onNewClient)
	return out
}

func (c *CueListsService) onNewClient(cl *client.Client) error {
	return errors.Join(
		cl.OnCueListCreated(func(s string, t *cues.CueListCreatedEvent) { c.app.Event.Emit(s, t) }),
		cl.OnCueListAttributesUpdated(func(s string, t *cues.CueListAttributesUpdatedEvent) { c.app.Event.Emit(s, t) }),
		cl.OnCueListDeleted(func(s string, t *cues.CueListDeletedEvent) { c.app.Event.Emit(s, t) }),
	)
}

func (c *CueListsService) CreateCueList(num float64, cueListType string) bool {
	err := c.clientManager.WithClient(func(c *client.Client) error {
		_, err := c.CreateCueList(num, cueListType)
		return err
	})

	if err != nil {
		c.logger.Error("failed to create cue list", "err", err)
		return false
	}

	return true
}

func (c *CueListsService) GetCueList(num float64) (*types.CueList, bool) {
	var out *types.CueList
	err := c.clientManager.WithClient(func(c *client.Client) error {
		md, err := c.GetCueListByNumber(uint(num))
		if err != nil {
			return err
		}
		out = md

		return nil
	})
	if err != nil {
		c.logger.Error("failed to get cue list", "err", err)
		return nil, false
	}
	return out, true
}

func (c *CueListsService) EnumerateCueLists() ([]types.CueList, bool) {
	var out []types.CueList
	err := c.clientManager.WithClient(func(c *client.Client) error {
		md, err := c.EnumerateCueLists()
		if err != nil {
			return err
		}
		out = md
		return nil
	})

	if err != nil {
		c.logger.Error("failed to enumerate cue lists", "err", err)
		return nil, false
	}

	return out, true
}

func (c *CueListsService) UpdateCueListAttributesField(id, field string, value interface{}) bool {
	err := c.clientManager.WithClient(func(c *client.Client) error {
		return c.UpdateCueListField(id, field, value)
	})
	if err != nil {
		c.logger.Error("failed to set cue list attributes field", "err", err)
		return false
	}

	return true
}

func (c *CueListsService) DeleteCueList(id string) bool {
	err := c.clientManager.WithClient(func(c *client.Client) error {
		return c.DeleteCueList(id)
	})
	if err != nil {
		c.logger.Error("failed to delete cue list", "err", err)
		return false
	}
	return true
}
