// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"log/slog"
	"os"

	"gitlab.com/stexxo/dynocue/internal/gui"
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
