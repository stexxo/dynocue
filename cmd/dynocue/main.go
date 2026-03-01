package main

import (
	"log/slog"
	"os"

	"gitlab.com/stexxo/dynocue/dynod/internal/gui"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)

	err := gui.NewGui().Run()
	if err != nil {
		slog.Error("GUI failed to start", "error", err)
	}

	slog.Info("DynoCue has shutdown")
}
