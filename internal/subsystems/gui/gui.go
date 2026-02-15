package gui

import (
	"fmt"
	"log/slog"
	"sync/atomic"

	"github.com/nats-io/nats.go"
	"github.com/wailsapp/wails/v3/pkg/application"
	"gitlab.com/stexxo/dynocue/dynod/frontend"
	"gitlab.com/stexxo/dynocue/dynod/internal/subsystems"
)

type Gui struct {
	app  *application.App
	conn *nats.Conn

	started    atomic.Bool
	guiRunning atomic.Bool
}

func NewGui() *Gui {
	return &Gui{}
}

func (g *Gui) Start(c *nats.Conn) error {
	if g.started.Load() {
		return subsystems.ErrStarted
	}

	g.conn = c

	g.app = application.New(application.Options{
		Name:        "gui",
		Description: "A demo of using raw HTML & CSS",
		Services:    []application.Service{},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(frontend.Assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: false,
		},
	})

	g.app.Window.New()

	g.started.Store(true)

	return nil
}

func (g *Gui) Stop() error {
	if g.guiRunning.Load() {
		g.app.Quit()
	}
	g.conn.Close()
	g.app = nil
	g.started.Store(false)
	return nil
}

func (g *Gui) Gui() {
	g.guiRunning.Store(true)
	err := g.app.Run()
	if err != nil {
		fmt.Println(err)
		slog.Error("GUI has shutdown", "error", err)
	}
	g.guiRunning.Store(false)
}

func (g *Gui) Name() string {
	return "gui"
}
