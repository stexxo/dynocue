package main

import (
	"log/slog"
	"os"

	"gitlab.com/stexxo/dynocue/dynod/internal"
	"gitlab.com/stexxo/dynocue/dynod/internal/subsystems/gui"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)

	defaultShow := ""
	if len(os.Args) > 1 {
		defaultShow = os.Args[1]
	}

	// Build Application SubSystems
	app := internal.NewAppManager(defaultShow, 4080)
	g := gui.NewGui()
	app.Register(g)

	// Start Application
	if err := app.Start(); err != nil {
		slog.Error("Failed to start application", "error", err)
		os.Exit(1)
	}

	// Ui Has to run on main thread - starting the ui after subsystems are started
	g.Gui()

	if err := app.Stop(); err != nil {
		slog.Error("Error during graceful shutdown", "error", err)
	}
}
