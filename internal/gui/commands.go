package gui

import "github.com/wailsapp/wails/v3/pkg/application"

type Commands struct {
	app *application.App
}

func NewCommands() *Commands {
	return &Commands{}
}

func (cmds *Commands) SetApplication(app *application.App) {
	cmds.app = app
}
