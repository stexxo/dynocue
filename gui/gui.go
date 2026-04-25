// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package gui

import (
	"github.com/stexxo/dynocue/client"
	"github.com/stexxo/dynocue/core/logging"
	"github.com/stexxo/dynocue/gui/frontend"
	"github.com/stexxo/dynocue/gui/services"
	"github.com/wailsapp/wails/v3/pkg/application"
)

type Gui struct {
	app           *application.App
	logger        logging.Logger
	clientManager *client.Manager
}

func NewGui(logger logging.Logger) *Gui {
	g := &Gui{
		clientManager: client.NewClientManager(logger),
		logger:        logger,
	}

	g.app = application.New(application.Options{
		Name: "DynoCue",
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(frontend.Assets),
		},
	})

	g.app.RegisterService(application.NewService(services.NewSelectorService(g.clientManager, g.app, g.logger)))
	g.app.RegisterService(application.NewService(services.NewCueListsService(g.clientManager, g.app, g.logger)))
	g.app.RegisterService(application.NewService(services.NewCuesService(g.clientManager, g.app, g.logger)))

	g.app.Window.NewWithOptions(application.WebviewWindowOptions{
		Width:     1280,
		Height:    720,
		MaxWidth:  5000,
		MaxHeight: 5000,
		Title:     "DynoCue",
		URL:       "/",
	}) // Default Window
	return g
}

func (g *Gui) Run() error {
	return g.app.Run()
}
