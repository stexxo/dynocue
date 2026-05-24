// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package client

import (
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/stexxo/dynocue/components/system"
	"github.com/stexxo/dynocue/core/logging"
	"github.com/stexxo/dynocue/core/messaging"
)

type Client struct {
	messenger   *messaging.Messenger
	persistence *system.PersistenceManager
}

func NewClient(clientName string, conn *nats.Conn, logger logging.Logger) (*Client, error) {
	js, err := jetstream.New(conn)
	if err != nil {
		return nil, err
	}

	c := &Client{
		messenger: messaging.NewMessenger(&messaging.MessengerCfg{
			Conn:   conn,
			Js:     js,
			Logger: logger,
		}),
	}

	persistence, err := system.RegisterWithPersistence(c.messenger, logger, "client-"+clientName, "", "")
	if err != nil {
		return nil, err
	}

	c.persistence = persistence

	return c, nil
}

type EventCallback[T any] func(string, *T)
