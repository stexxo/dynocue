// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package bus

import (
	"errors"
	"fmt"
	"time"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
)

// NewBus creates a new Nats server that is started and ready to go
func NewBus() (*server.Server, error) {
	opts := &server.Options{
		DontListen:    true,
		NoSigs:        true,            // We handle signals in our main app logic
		MaxPayload:    1024 * 1024 * 8, // 8MB limit
		PingInterval:  20 * time.Second,
		MaxPingsOut:   3,
		WriteDeadline: 10 * time.Second,
	}

	ns, err := server.NewServer(opts)
	if err != nil {
		return nil, fmt.Errorf("nats initialization failed: %w", err)
	}

	go ns.Start()

	if !ns.ReadyForConnections(5 * time.Second) {
		return nil, errors.New("nats server failed to become ready")
	}

	return ns, nil
}

// GetInProcessConn returns a connection to the internal bus given a NATS server
func GetInProcessConn(s *server.Server) (*nats.Conn, error) {
	if s == nil || !s.Running() {
		return nil, errors.New("cannot connect: bus is not started")
	}

	nc, err := nats.Connect("",
		nats.InProcessServer(s),
		nats.Name("Internal-Subsystem"),
		nats.MaxReconnects(-1), // Keep trying if internal pipe glitches
		nats.ReconnectWait(1*time.Second),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create in-process connection: %w", err)
	}

	return nc, nil
}
