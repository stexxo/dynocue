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

type CuesService struct {
	clientManager *client.Manager
	app           *application.App
	logger        logging.Logger
}

func NewCuesService(manager *client.Manager, app *application.App, logger logging.Logger) *CuesService {
	out := &CuesService{
		app:           app,
		logger:        logger,
		clientManager: manager,
	}
	out.clientManager.OnNewClient(out.onNewClient)
	return out
}

func (c *CuesService) onNewClient(cl *client.Client) error {
	return errors.Join(
		cl.OnCueCreated(func(s string, t *types.CueMetadata) { c.app.Event.Emit(s, t) }),
		cl.OnCueMetadataUpdated(func(s string, t *types.CueMetadata) { c.app.Event.Emit(s, t) }),
		cl.OnCueRenumber(func(s string, r *client.CueRenumberEvent) { c.app.Event.Emit(s, r) }),
		cl.OnCueDeleted(func(s string, e *cues.CueDeletedEvent) { c.app.Event.Emit(s, e) }),
	)
}

func (c *CuesService) CreateCue(cueListId string, cueNumber float64) bool {
	err := c.clientManager.WithClient(func(c *client.Client) error {
		_, err := c.CreateCue(cueListId, cueNumber)
		return err
	})

	if err != nil {
		c.logger.Error("failed to create cue", "err", err)
		return false
	}

	return true
}

func (c *CuesService) EnumerateCues(cueListId string) ([]types.CueMetadata, bool) {
	var out []types.CueMetadata
	err := c.clientManager.WithClient(func(c *client.Client) error {
		md, err := c.EnumerateCues(cueListId)
		if err != nil {
			return err
		}
		out = md
		return nil
	})

	if err != nil {
		c.logger.Error("failed to enumerate cues", "err", err)
		return nil, false
	}

	return out, true
}

func (c *CuesService) GetCueByNumber(cueListNumber float64, cueNumber float64) (*types.CueMetadata, bool) {
	var out *types.CueMetadata
	err := c.clientManager.WithClient(func(c *client.Client) error {
		md, err := c.GetCueByNumber(cueListNumber, cueNumber)
		if err != nil {
			return err
		}
		out = md
		return nil
	})

	if err != nil {
		c.logger.Error("failed to get cue by number", "err", err)
		return nil, false
	}

	return out, true
}

func (c *CuesService) GetCueById(cueListId string, cueId string) (*types.CueMetadata, bool) {
	var out *types.CueMetadata
	err := c.clientManager.WithClient(func(c *client.Client) error {
		md, err := c.GetCueById(cueListId, cueId)
		if err != nil {
			return err
		}
		out = md
		return nil
	})

	if err != nil {
		c.logger.Error("failed to get cue by id", "err", err)
		return nil, false
	}

	return out, true
}

func (c *CuesService) UpdateCueMetadata(cueListId string, cueId string, field string, value any) bool {
	err := c.clientManager.WithClient(func(c *client.Client) error {
		_, err := c.UpdateCueMetadata(cueListId, cueId, field, value)
		return err
	})

	if err != nil {
		c.logger.Error("failed to update cue metadata", "err", err)
		return false
	}

	return true
}

func (c *CuesService) RenumberCue(cueListId string, cueId string, newNumber float64) bool {
	err := c.clientManager.WithClient(func(c *client.Client) error {
		return c.RenumberCue(cueListId, cueId, newNumber)
	})

	if err != nil {
		c.logger.Error("failed to renumber cue", "err", err)
		return false
	}

	return true
}

func (c *CuesService) DeleteCue(cueListId string, cueId string) bool {
	err := c.clientManager.WithClient(func(c *client.Client) error {
		return c.DeleteCue(cueListId, cueId)
	})

	if err != nil {
		c.logger.Error("failed to delete cue", "err", err)
		return false
	}

	return true
}
