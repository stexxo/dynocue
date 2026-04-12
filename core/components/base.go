package components

import (
	"sync/atomic"

	"github.com/go-playground/validator/v10"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/stexxo/dynocue/core/logging"
	"github.com/stexxo/dynocue/core/messaging"
)

type BaseComponent struct {
	started   atomic.Bool
	messenger *messaging.Messenger
	logger    logging.Logger

	name    string
	startFn func() error
}

func NewBaseComponent(name string, logger logging.Logger, onStart func() error) *BaseComponent {
	return &BaseComponent{
		logger:  logger,
		name:    name,
		startFn: onStart,
	}
}

func (b *BaseComponent) Start(conn *nats.Conn) error {
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

	b.logger.Debug("started subsystem " + b.name)

	b.started.Store(true)
	return nil
}

func (b *BaseComponent) Messenger() *messaging.Messenger {
	return b.messenger
}

func (b *BaseComponent) Logger() logging.Logger {
	return b.logger
}

func (b *BaseComponent) Name() string {
	return b.name
}

func (b *BaseComponent) Stop() error {
	if !b.started.Load() {
		return nil
	}
	b.Messenger().Conn().Close()
	b.started.Store(false)
	return nil
}
