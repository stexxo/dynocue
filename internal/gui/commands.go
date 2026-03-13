package gui

import (
	"strings"

	"github.com/nats-io/nats.go"
	"github.com/wailsapp/wails/v3/pkg/application"
	"gitlab.com/stexxo/dynocue/internal/show"
)

type Commands struct {
	app  *application.App
	show *show.Show
	conn *nats.Conn
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
	cmds.conn, err = s.GetConn()
	if err != nil {
		return "", false
	}

	return path, true
}

func (cmds *Commands) CloseShow() {
	if cmds.show != nil {
		cmds.conn.Close()
		cmds.show.Close()
		cmds.show = nil
		cmds.conn = nil
	}
}
