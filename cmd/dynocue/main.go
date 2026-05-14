// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"log/slog"
	"os"

	"github.com/stexxo/dynocue/gui"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: true}))
	slog.SetDefault(logger)

	// Build and Run GUI
	g := gui.NewGui(logger)
	err := g.Run()
	if err != nil {
		slog.Error("Failed to start DynoCue GUI", "error", err)
	}

	slog.Info("DynoCue has shutdown")
}
