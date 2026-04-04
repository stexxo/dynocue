package core

import (
	"cmp"
	"fmt"
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
	s, err := server.NewServer(&server.Options{})
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

	d.server.Start()

	return nil
}
