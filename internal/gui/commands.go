package gui

import (
	"strings"

	"github.com/wailsapp/wails/v3/pkg/application"
	"gitlab.com/stexxo/dynocue/internal/show"
)

type Commands struct {
	app  *application.App
	show *show.Show
}

func NewCommands() *Commands {
	return &Commands{}
}

func (cmds *Commands) SetApplication(app *application.App) {
	cmds.app = app
}

func (cmds *Commands) OpenShow(path string) (string, bool) {
	if !strings.HasSuffix(path, ".dynocue") {
		path = path + ".dynocue"
	}

	if cmds.show != nil {
		cmds.show.Close()
	}
	s, err := show.NewShow(path)
	if err != nil {
		return "", false
	}
	cmds.show = s
	return path, true
}

func (cmds *Commands) CloseShow() {
	if cmds.show != nil {
		cmds.show.Close()
		cmds.show = nil
	}
}
