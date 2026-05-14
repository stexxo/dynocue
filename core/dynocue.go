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

	cfg.Logger.Debug(fmt.Sprintf("dynocue initialized with %d subsystems", len(subs)))

	return &DynoCue{
		server:    s,
		subsystem: subs,
		logger:    cmp.Or[logging.Logger](cfg.Logger, logging.NewNoopLogger())}, nil
}

func (d *DynoCue) Start() error {
	if d.started.Load() {
		return nil
	}

	d.logger.Debug("starting nats")
	d.server.Start()
	for range 10 {
		if !d.server.Running() {
			time.Sleep(50 * time.Millisecond)
		}
		break
	}

	if !d.server.Running() {
		d.logger.Error("dynocue failed to start due to nats server startup timing out")
		return fmt.Errorf("dynocue failed to start nats server")
	}

	d.logger.Debug("nats server started")

	d.logger.Debug("starting subsystems")
	execgroup := errgroup.Group{}
	for _, sub := range d.subsystem {
		execgroup.Go(func() error {
			d.logger.Debug(fmt.Sprintf("starting subsystem %s", sub.Name()))
			nc, err := d.GetInProcessConn(sub.Name())
			if err != nil {
				d.logger.Error(fmt.Sprintf("failed to create in process connection to nats server for subsystem %s", sub.Name()), "error", err)
				return fmt.Errorf("\nfailed to create in process connection to nats server for subsystem %s: %w\n", sub.Name(), err)
			}

			err = sub.Start(nc)
			if err != nil {
				return fmt.Errorf("failed to start subsystem %s: %w", sub.Name(), err)
			}

			d.logger.Debug(fmt.Sprintf("started subsystem %s", sub.Name()))
			return nil
		})
	}

	err := execgroup.Wait()
	if err != nil {
		d.logger.Error("failed to start subsystems", "error", err)
		return err
	}

	d.logger.Debug("subsystems started")

	d.started.Store(true)

	d.logger.Debug("dynocue started")

	return nil
}

func (d *DynoCue) GetInProcessConn(name string) (*nats.Conn, error) {
	return nats.Connect("", nats.InProcessServer(d.server), nats.MaxReconnects(-1), nats.Name(name), nats.ReconnectWait(1*time.Second))
}
