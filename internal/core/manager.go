// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package core

import (
	"errors"
	"fmt"
	"os"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	ibus "gitlab.com/stexxo/dynocue/internal/bus"
	"gitlab.com/stexxo/dynocue/internal/core/cues"
	"go.etcd.io/bbolt"
)

// Subsystem defines an interface for all core components that require
// lifecycle management and cleanup.
type Subsystem interface {
	Close() error
}

// Show represents a core session, managing the database, message bus,
// and all active subsystems.
type Show struct {
	db       *bbolt.DB
	bus      *server.Server
	savePath string

	subsystem []Subsystem
}

// NewShow initializes a new core session at the given directory path,
// starting the message bus, database, and all required subsystems.
func NewShow(path string) (s *Show, err error) {
	fi, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(path, 0755); err != nil {
				return nil, fmt.Errorf("failed to create directory %s: %w", path, err)
			}
		} else {
			return nil, fmt.Errorf("failed to stat path %s: %w", path, err)
		}
	} else if !fi.IsDir() {
		return nil, fmt.Errorf("path %s exists and is not a directory", path)
	}

	b, err := ibus.NewBus()
	if err != nil {
		return nil, err
	}
	db, err := bbolt.Open(path+"/dynocue.db", 0600, &bbolt.Options{})
	if err != nil {
		return nil, err
	}
	s = &Show{
		db:       db,
		savePath: path,
		bus:      b,
	}
	defer func() {
		if err != nil {
			err = errors.Join(err, s.Close())
		}
	}()

	// Build Subsystems Required for Show

	// CueSystem
	conn, err := ibus.GetInProcessConn(b)
	if err != nil {
		return nil, err
	}
	c, err := cues.NewCues(conn, db)
	if err != nil {
		return nil, err
	}
	s.subsystem = append(s.subsystem, c)

	return
}

// GetConn returns a NATS connection for communicating with the core's internal bus.
func (s *Show) GetConn() (*nats.Conn, error) {
	return ibus.GetInProcessConn(s.bus)
}

// Close shuts down the core session, closing all subsystems, the message bus,
// and the database.
func (s *Show) Close() error {
	for _, subsystem := range s.subsystem {
		if err := subsystem.Close(); err != nil {
			return err
		}
	}

	s.bus.Shutdown()
	return s.db.Close()
}
