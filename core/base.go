package core

import (
	"sync/atomic"

	"github.com/go-playground/validator/v10"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/stexxo/dynocue/core/logging"
	"github.com/stexxo/dynocue/core/messaging"
)

type SubsystemCore struct {
	started   atomic.Bool
	messenger *messaging.Messenger
	logger    logging.Logger

	name    string
	startFn func() error
}

func NewSubsystemCore(name string, logger logging.Logger, onStart func() error) *SubsystemCore {
	return &SubsystemCore{
		logger:  logger,
		name:    name,
		startFn: onStart,
	}
}

func (b *SubsystemCore) Start(conn *nats.Conn) error {
	if b.started.Load() {
		return nil
	}
	js, err := jetstream.New(conn)
	if err != nil {
		return err
	}
	b.messenger = messaging.NewMessenger(&messaging.MessengerCfg{Conn: conn, Logger: b.logger, Validator: validator.New(), Js: js})

	err = b.startFn()
	if err != nil {
		b.logger.Error("failed to start subsystem", "error", err)
		return err
	}

	b.logger.Debug("subsystem  " + b.name + " has started successfully")

	b.started.Store(true)
	return nil
}

func (b *SubsystemCore) Messenger() *messaging.Messenger {
	return b.messenger
}

func (b *SubsystemCore) Logger() logging.Logger {
	return b.logger
}

func (b *SubsystemCore) Name() string {
	return b.name
}

func (b *SubsystemCore) Stop() error {
	if !b.started.Load() {
		return nil
	}
	b.Messenger().Conn().Close()
	b.started.Store(false)
	return nil
}
