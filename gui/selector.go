// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package gui

import (
	"errors"

	"github.com/stexxo/dynocue/client"
	"github.com/stexxo/dynocue/components/cues"
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

func (s *Selector) localConn() error {
	c, err := core.NewDynoCue(&core.Config{
		Subsystems: []core.Subsystem{
			system.NewPersistence(s.logger),
			show.New(s.logger),
			cues.New(s.logger),
		},
		Logger: s.logger,
	})
	if err != nil {
		s.logger.Error("Failed to initialize DynoCue: ", "err", err)
		return err
	}

	s.logger.Debug("starting dynocue core")
	err = c.Start()
	if err != nil {
		s.logger.Error("Failed to start DynoCue: ", "err", err)
		return err
	}

	s.logger.Debug("connecting to dynocue core in process")
	err = s.clientManager.ConnectLocal(c)
	if err != nil {
		s.logger.Error("Failed to connect to DynoCue: ", "err", err)
		return err
	}

	return nil
}

func (s *Selector) NewShow() bool {
	s.logger.Debug("creating new local show")

	if !s.clientManager.Connected() || s.clientManager.Remote() {
		s.logger.Debug("initializing dynocue core")

		err := s.localConn()
		if err != nil {
			s.logger.Warn("Failed to initialize DynoCue core: ", "err", err)
			return false
		}
	}

	err := s.clientManager.WithClient(func(c *client.Client) error {
		return c.NewShow()
	})
	if err != nil {
		return false
	}

	s.logger.Debug("navigating windows to dashboard")
	for _, w := range s.app.Window.GetAll() {
		w.SetURL("/show/dashboard")
	}

	return true

}

func (s *Selector) saveDialog() (string, error) {
	dia := s.app.Dialog.SaveFileWithOptions(&application.SaveFileDialogOptions{
		Title: "Save Cueing",
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

func (s *Selector) OpenShow() bool {
	dia := s.app.Dialog.OpenFileWithOptions(&application.OpenFileDialogOptions{})
	res, err := dia.PromptForSingleSelection()
	if err != nil {
		s.logger.Error("open show failed", "err", err.Error())
		return true
	}

	if res == "" {
		return true
	}

	if !s.clientManager.Connected() || s.clientManager.Remote() {
		s.logger.Debug("initializing dynocue core")

		err := s.localConn()
		if err != nil {
			s.logger.Warn("Failed to initialize DynoCue core: ", "err", err)
			return false
		}
	}

	s.logger.Debug("opening local show")
	err = s.clientManager.WithClient(func(c *client.Client) error {
		return c.OpenShow(res)
	})

	if err != nil {
		s.logger.Warn("Failed to open show ", err)
		return false
	}

	s.logger.Debug("navigating windows to dashboard")
	for _, w := range s.app.Window.GetAll() {
		w.SetURL("/show/dashboard")
	}

	s.logger.Debug("opened show successfully")
	return true
}

func (s *Selector) CloseShow() bool {
	s.logger.Debug("closing show")
	s.logger.Debug("navigating windows to dashboard")
	for _, w := range s.app.Window.GetAll() {
		w.SetURL("/")
	}

	err := s.clientManager.Disconnect()
	if err != nil {
		s.logger.Warn("Failed to close DynoCue core: ", "err", err)
		return false
	}

	return true
}
