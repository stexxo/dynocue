package main

import (
	"log/slog"
	"os"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)

	err := NewGui().Run()
	if err != nil {
		slog.Error("GUI failed to start", "error", err)
	}

	slog.Info("DynoCue has shutdown")
}
