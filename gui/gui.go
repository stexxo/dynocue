// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package gui

import (
	"github.com/stexxo/dynocue/core"
	"github.com/stexxo/dynocue/gui/frontend"
	"github.com/wailsapp/wails/v3/pkg/application"
)

type Gui struct {
	app *application.App
}

func NewGui(core *core.DynoCue) *Gui {
	g := &Gui{}

	g.app = application.New(application.Options{
		Name:     "DynoCue",
		Services: []application.Service{},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(frontend.Assets),
		},
	})

	g.app.Window.NewWithOptions(application.WebviewWindowOptions{
		Width:     1280,
		Height:    720,
		MinWidth:  800,
		MinHeight: 600,
		Title:     "DynoCue",
		URL:       "/",
	}) // Default Window

	return g
}

func (g *Gui) Run() error {
	return g.app.Run()
}
