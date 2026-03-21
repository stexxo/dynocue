// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package gui

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/nats-io/nats.go"
	"github.com/stexxo/dynocue/internal/core"
	apibus "github.com/stexxo/dynocue/pkg/bus"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// Commands handles backend operations exposed to the frontend,
// including core management and communication with the NATS bus.
type Commands struct {
	app *application.App

	show          *core.Show
	conn          *nats.Conn
	subscriptions []*nats.Subscription
}

// NewCommands creates a new instance of Commands.
func NewCommands() *Commands {
	return &Commands{}
}

// SetApplication sets the Wails application instance for the commands.
func (c *Commands) SetApplication(app *application.App) {
	c.app = app
}

// OpenShow opens a core file at the given path, initializing the core
// system and subscribing to relevant events.
func (c *Commands) OpenShow(path string) (string, bool) {
	if path == "" {
		return "", false
	}

	if !strings.HasSuffix(path, ".dynocue") {
		path = path + ".dynocue"
	}

	if c.show != nil {
		c.show.Close()
	}
	s, err := core.NewShow(path)
	if err != nil {
		return "", false
	}
	c.show = s
	c.conn, err = s.GetConn()
	if err != nil {
		return "", false
	}

	err = c.SubscribeToAll()
	if err != nil {
		slog.Error("Failed to subscribe to all events", "error", err)
		return "", false
	}

	return path, true
}

// CloseShow closes the currently open core and cleans up resources.
func (c *Commands) CloseShow() {
	if c.show != nil {
		c.conn.Close()
		c.show.Close()
		c.show = nil
		c.conn = nil
	}
}

func makeRequest[T any, E any](c *Commands, subject string, input T) (*E, error) {
	if c.show == nil || c.conn == nil {
		return nil, errors.New("core closed")
	}

	res, err := apibus.Request[T, E](c.conn, subject, input)
	if err != nil {
		return nil, err
	}

	if res == nil {
		return nil, errors.New("no response returned")
	}

	if res.MessageError != nil {
		return nil, fmt.Errorf("error in response: %s", res.MessageError.ErrorMessage)
	}

	return res.ResponseValue, nil
}
