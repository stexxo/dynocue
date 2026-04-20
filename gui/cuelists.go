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

func (c *CueListsService) onNewClient(client *client.Client) error {
	return errors.Join(
		client.OnCueListCreated(func(s string, t *types.CueListMetadata) { c.app.Event.Emit(s, t) }),
	)
}

func (c *CueListsService) CreateCueList(num float64) bool {
	err := c.clientManager.WithClient(func(c *client.Client) error {
		_, err := c.CreateCueList(num)
		return err
	})

	if err != nil {
		c.logger.Error("failed to create cue list", "err", err)
		return false
	}

	return true
}

func (c *CueListsService) GetCueList(num float64) (*types.CueListMetadata, error) {
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
		return nil, err
	}
	return out, nil
}
