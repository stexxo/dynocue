// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package core

import (
	"cmp"
	"fmt"
	"os"
	"sync/atomic"
	"time"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"github.com/stexxo/dynocue/core/logging"
	"golang.org/x/sync/errgroup"
)

type Subsystem interface {
	Start(conn *nats.Conn) error
	Name() string
	Stop() error
}

type Config struct {
	Subsystems []Subsystem
	Logger     logging.Logger
}

type DynoCue struct {
	server    *server.Server
	subsystem []Subsystem
	started   atomic.Bool
	logger    logging.Logger
}

func NewDynoCue(cfg *Config) (*DynoCue, error) {
	s, err := server.NewServer(&server.Options{
		JetStream:  true,
		ServerName: "DynoCue",
		StoreDir:   os.TempDir() + "/dynocue",
		MaxPayload: 1024 * 1024 * 64,
	})
	if err != nil {
		return nil, err
	}

	subs := make([]Subsystem, len(cfg.Subsystems))
	copy(subs, cfg.Subsystems)

	return &DynoCue{
		server:    s,
		subsystem: subs,
		logger:    cmp.Or[logging.Logger](cfg.Logger, logging.NewNoopLogger())}, nil
}

func (d *DynoCue) Start() error {
	if d.started.Load() {
		return nil
	}

	d.server.Start()
	for range 10 {
		if !d.server.Running() {
			time.Sleep(50 * time.Millisecond)
		}
		break
	}

	execgroup := errgroup.Group{}
	for _, sub := range d.subsystem {
		execgroup.Go(func() error {
			nc, err := nats.Connect("", nats.InProcessServer(d.server), nats.MaxReconnects(-1), nats.Name(sub.Name()), nats.ReconnectWait(1*time.Second))
			if err != nil {
				return fmt.Errorf("\nfailed to create in process connection to nats server for subsystem %s: %w\n", sub.Name(), err)
			}
			return sub.Start(nc)
		})
	}

	err := execgroup.Wait()
	if err != nil {
		d.logger.Error("failed to start subsystems", "error", err)
		return err
	}

	d.started.Store(true)

	return nil
}
