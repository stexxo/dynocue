package subsystems

import (
	"errors"

	"gitlab.com/stexxo/dynocue/dynod/internal/bus"
	"gitlab.com/stexxo/dynocue/dynod/internal/show"
)

var ErrStarted = errors.New("subsystem already started")
var ErrStopped = errors.New("subsystem already stopped")

type Subsystem interface {
	Name() string
	Start(bus *bus.Client, show *show.Show) error
	Stop() error
}
