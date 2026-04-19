package gui

import (
	"github.com/stexxo/dynocue/client"
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

func (service *CueListsService) onNewClient(client *client.Client) error {
	return nil
}
