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
	"github.com/wailsapp/wails/v3/pkg/events"
)

type Gui struct {
	app           *application.App
	logger        logging.Logger
	clientManager *client.Manager
}

func NewGui(logger logging.Logger) *Gui {
	g := &Gui{
		clientManager: client.NewClientManager("GUI", logger),
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
	g.app.RegisterService(application.NewService(services.NewExecutionService(g.clientManager, g.app, g.logger)))
	g.app.RegisterService(application.NewService(services.NewActionsService(g.clientManager, g.app, g.logger)))
	g.app.RegisterService(application.NewService(services.NewActionTemplatesService(g.clientManager, g.app, g.logger)))
	g.app.RegisterService(application.NewService(services.NewAudioService(g.clientManager, g.app, g.logger)))

	win := g.app.Window.NewWithOptions(application.WebviewWindowOptions{
		Width:          1280,
		Height:         720,
		MaxWidth:       5000,
		MaxHeight:      5000,
		Title:          "DynoCue",
		URL:            "/",
		EnableFileDrop: true,
	}) // Default Window

	win.OnWindowEvent(events.Common.WindowFilesDropped, func(event *application.WindowEvent) {
		files := event.Context().DroppedFiles()
		details := event.Context().DropTargetDetails()

		for _, file := range files {
			g.app.Event.Emit(details.ElementID, map[string]string{
				"file": file,
			})
		}
	})
	return g
}

func (g *Gui) Run() error {
	return g.app.Run()
}
