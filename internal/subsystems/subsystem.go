package subsystems

import (
	"errors"

	"github.com/nats-io/nats.go"
)

var ErrStarted = errors.New("subsystem already started")
var ErrStopped = errors.New("subsystem already stopped")

type Subsystem interface {
	Name() string
	Start(*nats.Conn) error
	Stop() error
}
