package gui

import (
	"errors"

	"github.com/stexxo/dynocue/client"
	"github.com/stexxo/dynocue/components/cues/types"
	"github.com/stexxo/dynocue/core/logging"
	"github.com/wailsapp/wails/v3/pkg/application"
)

type CueListsService struct {
	clientManager *ClientManager
	app           *application.App
	logger        logging.Logger
}

func NewCueListsService(manager *ClientManager, app *application.App, logger logging.Logger) *CueListsService {
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
		cl.OnCueListCreated(func(s string, t *types.CueListMetadata) { c.app.Event.Emit(s, t) }),
		cl.OnCueListMetadataUpdated(func(s string, t *types.CueListMetadata) { c.app.Event.Emit(s, t) }),
		cl.OnCueListRenumber(func(s string, r *client.RenumberEvent) { c.app.Event.Emit(s, r) }),
		cl.OnCueListDeleted(func(s string, f *float64) { c.app.Event.Emit(s, *f) }),
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

func (c *CueListsService) GetCueList(num float64) (*types.CueListMetadata, bool) {
	var out *types.CueListMetadata
	err := c.clientManager.WithClient(func(c *client.Client) error {
		md, err := c.GetCueList(num)
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

func (c *CueListsService) EnumerateCueLists() ([]types.CueListMetadata, bool) {
	var out []types.CueListMetadata
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

func (c *CueListsService) SetCueListLabel(num float64, label string) bool {
	err := c.clientManager.WithClient(func(c *client.Client) error {
		_, err := c.SetCueListLabel(num, label)
		return err
	})
	if err != nil {
		c.logger.Error("failed to set cue list label", "err", err)
		return false
	}

	return true
}

func (c *CueListsService) RenumberCueList(origNum, newNum float64) bool {
	err := c.clientManager.WithClient(func(c *client.Client) error {
		return c.RenumberCueList(origNum, newNum)
	})
	if err != nil {
		c.logger.Error("failed to renumber cue list", "err", err)
		return false
	}

	return true
}

func (c *CueListsService) DeleteCueList(num float64) bool {
	err := c.clientManager.WithClient(func(c *client.Client) error {
		return c.DeleteCueList(num)
	})
	if err != nil {
		c.logger.Error("failed to delete cue list", "err", err)
		return false
	}
	return true
}
