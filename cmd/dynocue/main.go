// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"log/slog"
	"os"

	"github.com/stexxo/dynocue/core"
	"github.com/stexxo/dynocue/gui"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)

	// Build Core
	d, err := core.NewDynoCue(&core.Config{
		Logger:     logger,
		Subsystems: []core.Subsystem{},
	})
	if err != nil {
		slog.Error("Failed to start DynoCue core", "error", err)
		return
	}

	// Start Core
	err = d.Start()
	if err != nil {
		slog.Error("Failed to start DynoCue core", "error", err)
		return
	}

	// Build and Run GUI
	g := gui.NewGui(d)
	err = g.Run()
	if err != nil {
		slog.Error("Failed to start DynoCue core", "error", err)
	}

	slog.Info("DynoCue has shutdown")
}
