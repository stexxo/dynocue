// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package client

import (
	"github.com/nats-io/nats.go"
	"github.com/stexxo/dynocue/core/logging"
	"github.com/stexxo/dynocue/core/messaging"
)

type Client struct {
	messenger *messaging.Messenger
}

func NewClient(conn *nats.Conn, logger logging.Logger) *Client {
	return &Client{
		messenger: messaging.NewMessenger(&messaging.MessengerCfg{
			Conn:   conn,
			Logger: logger,
		}),
	}
}

type EventCallback[T any] func(string, *T)
