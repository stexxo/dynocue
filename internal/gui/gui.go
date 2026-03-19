// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package gui

import (
	"github.com/wailsapp/wails/v3/pkg/application"
	"gitlab.com/stexxo/dynocue/frontend"
)

type Gui struct {
	app *application.App
}

func NewGui() *Gui {
	g := &Gui{}

	cmds := NewCommands()
	g.app = application.New(application.Options{
		Name: "DynoCue",
		Services: []application.Service{
			application.NewService(cmds),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(frontend.Assets),
		},
	})

	cmds.SetApplication(g.app)

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
