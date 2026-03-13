package main

import (
	"github.com/wailsapp/wails/v3/pkg/application"
	"gitlab.com/stexxo/dynocue/frontend"
)

type Gui struct {
	app *application.App
}

func NewGui() *Gui {
	g := &Gui{}

	g.app = application.New(application.Options{
		Name:     "DynoCue",
		Services: []application.Service{},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(frontend.Assets),
		},
	})

	g.app.Window.NewWithOptions(application.WebviewWindowOptions{
		Width:  1280,
		Height: 720,
		Title:  "DynoCue",
		URL:    "/",
	}) // Default Window

	return g
}

func (g *Gui) Run() error {
	return g.app.Run()
}
