package gui

import (
	"log/slog"
	"strings"

	"github.com/google/uuid"
	"github.com/wailsapp/wails/v3/pkg/application"
	"gitlab.com/stexxo/dynocue/dynod/internal/bus"
)

func (g *Gui) CreateWindow(location, title string) {
	if g.app == nil {
		return
	}

	windowName := uuid.NewString()

	title = "DynoCue"
	location = "/"
	if g.show.IsInitialized() {
		location = "/show"
	}

	g.app.Window.NewWithOptions(application.WebviewWindowOptions{
		Name:             windowName,
		Title:            title,
		BackgroundColour: application.NewRGB(27, 38, 54),
		URL:              location,
	})

}

func (g *Gui) OpenShow(msg bus.Message) {
	slog.Debug("received request to open show")

	path, err := g.app.Dialog.OpenFileWithOptions(&application.OpenFileDialogOptions{
		CanCreateDirectories: false,
		CanChooseFiles:       false,
		CanChooseDirectories: true,
		Title:                "Open Show",
		ButtonText:           "Open",
	}).PromptForMultipleSelection()
	if err != nil || len(path) == 0 {
		slog.Error("failed to open open dialog for show", "error", err)
		g.evMgr.RespondHelper(msg, []byte("FAILED"))
		return
	}

	if !strings.HasSuffix(path[0], ".dynocue") {
		slog.Warn("invalid file extension")
		g.evMgr.RespondHelper(msg, []byte("INVALID FILE EXTENSION"))
		return
	}

	resp, ok := g.evMgr.RequestHelper("show.load", []byte(path[0]))
	if !ok || string(resp.Data) != "SUCCESS" {
		slog.Error("failed to load show", "error", err)
		g.evMgr.RespondHelper(msg, []byte("FAILED"))
		return
	}

	for _, w := range g.app.Window.GetAll() {
		w.SetURL("/show/cues")
	}

	g.evMgr.RespondHelper(msg, []byte("SUCCESS"))
}

func (g *Gui) CreateShow(msg bus.Message) {
	slog.Debug("received request for new show")

	path, err := g.app.Dialog.SaveFileWithOptions(&application.SaveFileDialogOptions{
		Title:                "New Show",
		ButtonText:           "Create",
		CanCreateDirectories: true,
	}).PromptForSingleSelection()
	if err != nil || path == "" {
		slog.Error("failed to open save dialog for new show", "error", err)
		g.evMgr.RespondHelper(msg, []byte("FAILED"))
		return
	}

	if !strings.HasSuffix(path, ".dynocue") {
		path = path + ".dynocue"
	}

	resp, ok := g.evMgr.RequestHelper("show.new", []byte(path))
	if !ok || string(resp.Data) != "SUCCESS" {
		slog.Error("failed to create new show", "error", err)
		g.evMgr.RespondHelper(msg, []byte("FAILED"))
		return
	}

	for _, w := range g.app.Window.GetAll() {
		w.SetURL("/show/cues")
	}

	g.evMgr.RespondHelper(msg, []byte("SUCCESS"))
}

func (g *Gui) CloseShow(msg bus.Message) {
	slog.Debug("received request to close show from window " + string(msg.Data))
	g.evMgr.RequestHelper("show.close", nil)

	w, ok := g.app.Window.Get(string(msg.Data))
	if ok {
		w.Focus()
		w.SetURL("/")
		if len(g.app.Window.GetAll()) > 1 {
			for _, w := range g.app.Window.GetAll() {
				if w.Name() != w.Name() {
					w.Close()
				}
			}
		}
	}

	g.evMgr.RespondHelper(msg, []byte("SUCCESS"))
}
