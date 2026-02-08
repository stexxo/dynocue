package gui

import (
	"errors"
	"fmt"
	"log/slog"
	"sync/atomic"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
	"gitlab.com/stexxo/dynocue/dynod/frontend"
	"gitlab.com/stexxo/dynocue/dynod/internal/bus"
	"gitlab.com/stexxo/dynocue/dynod/internal/show"
	"gitlab.com/stexxo/dynocue/dynod/internal/subsystems"
	"gitlab.com/stexxo/dynocue/dynod/internal/subsystems/gui/api"
)

type Gui struct {
	evMgr *bus.Client
	show  *show.Show
	app   *application.App

	// services
	lifecycleService *api.LifecycleService

	started atomic.Bool
	running atomic.Bool
}

func NewGui() *Gui {
	return &Gui{}
}

func (g *Gui) Start(client *bus.Client, show *show.Show) error {
	if g.started.Load() {
		return subsystems.ErrStarted
	}

	g.evMgr = client
	g.show = show

	g.evMgr.Subscribe("gui.show.new", g.CreateShow)
	g.evMgr.Subscribe("gui.show.load", g.OpenShow)
	g.evMgr.Subscribe("gui.show.close", g.CloseShow)

	g.app = application.New(application.Options{
		Name:        "gui",
		Description: "A demo of using raw HTML & CSS",
		Services: []application.Service{
			application.NewService(api.NewLifecycleService(g.evMgr)),
			application.NewService(api.NewCueService(g.evMgr)),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(frontend.Assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: false,
		},
	})

	g.evMgr.Subscribe("show.>", func(m bus.Message) {
		slog.Debug("emitting event to gui", "subject", m.Subject)
		g.app.Event.Emit(m.Subject, m.Data)
	})

	g.CreateWindow("/", "DynoCue")

	g.app.Event.OnApplicationEvent(events.Common.ApplicationStarted, func(event *application.ApplicationEvent) {
		if g.show.IsInitialized() {
			g.evMgr.SendHelper("gui.window.main.new", nil)
		} else {
			g.evMgr.SendHelper("gui.window.splash.new", nil)
		}
	})
	g.started.Store(true)

	return nil
}

func (g *Gui) Stop() error {
	if !g.started.Load() {
		return subsystems.ErrStopped
	}
	err := errors.Join(g.evMgr.Stop())
	if err == nil {
		g.evMgr = nil
	}
	if g.running.Load() {
		g.app.Quit()
	}
	g.show = nil
	g.app = nil
	g.started.Store(false)
	return err
}

func (g *Gui) Run() {
	g.running.Store(true)
	err := g.app.Run()
	if err != nil {
		fmt.Println(err)
		slog.Error("GUI has shutdown", "error", err)
	}
	g.running.Store(false)
}

func (g *Gui) Name() string {
	return "gui"
}
