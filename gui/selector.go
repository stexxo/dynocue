package gui

import (
	"errors"

	"github.com/stexxo/dynocue/client"
	"github.com/stexxo/dynocue/components/show"
	"github.com/stexxo/dynocue/components/system"
	"github.com/stexxo/dynocue/core"
	"github.com/stexxo/dynocue/core/logging"
	"github.com/wailsapp/wails/v3/pkg/application"
)

type Selector struct {
	clientManager *ClientManager
	app           *application.App
	logger        logging.Logger
}

func NewSelector(manager *ClientManager, app *application.App, logger logging.Logger) *Selector {
	return &Selector{
		clientManager: manager,
		app:           app,
		logger:        logger,
	}
}

func (s *Selector) NewShow() bool {
	s.logger.Debug("creating new local show")

	s.logger.Debug("initializing dynocue core")
	c, err := core.NewDynoCue(&core.Config{
		Subsystems: []core.Subsystem{
			system.NewPersistence(s.logger),
			show.NewShow(s.logger),
		},
		Logger: s.logger,
	})
	if err != nil {
		s.logger.Warn("Failed to initialize DynoCue: ", err)
		return false
	}

	s.logger.Debug("starting dynocue core")
	err = c.Start()
	if err != nil {
		s.logger.Warn("Failed to start DynoCue: ", err)
		return false
	}

	s.logger.Debug("connecting to dynocue core in process")
	err = s.clientManager.ConnectLocal(c)
	if err != nil {
		s.logger.Warn("Failed to connect to DynoCue: ", err)
		return false
	}

	s.logger.Debug("navigating windows to dashboard")
	for _, w := range s.app.Window.GetAll() {
		w.SetURL("/show/dashboard")
	}

	s.logger.Debug("new show initialized successfully")
	return true
}

func (s *Selector) saveDialog() (string, error) {
	dia := s.app.Dialog.SaveFileWithOptions(&application.SaveFileDialogOptions{
		Title: "Save Show",
	})
	return dia.PromptForSingleSelection()
}

func (s *Selector) SaveShow() bool {
	s.logger.Debug("saving local show")

	err := s.clientManager.WithClient(func(c *client.Client) error {
		err := c.SaveShow("")
		if errors.Is(err, client.NoSaveLocation) {
			s.logger.Debug("no save location, prompting for one")
			res, err := s.saveDialog()
			if err != nil {
				s.logger.Error("save dialog failed", "err", err.Error())
				return err
			}
			if res == "" {
				s.logger.Debug("no save location provided, exiting")
				return nil
			}

			return c.SaveShow(res)
		}

		return err
	})
	if err != nil {
		s.logger.Warn("Failed to save DynoCue: ", "err", err.Error())
		return false
	}

	s.logger.Debug("saved show successfully")
	return true
}

func (s *Selector) SaveShowAs() bool {
	s.logger.Debug("saving local show as")

	err := s.clientManager.WithClient(func(c *client.Client) error {
		res, err := s.saveDialog()
		if err != nil {
			s.logger.Error("save dialog failed", "err", err.Error())
			return err
		}
		if res == "" {
			s.logger.Debug("no save location provided, exiting")
			return nil
		}

		return c.SaveShow(res)
	})
	if err != nil {
		s.logger.Warn("Failed to save DynoCue: ", err)
		return false
	}

	s.logger.Debug("saved show successfully")

	return true
}
