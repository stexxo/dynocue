package subsystems

import (
	"errors"
	"log/slog"
	"os"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"gitlab.com/stexxo/dynocue/internal/bus"
	"gitlab.com/stexxo/dynocue/internal/subsystems/database"
)

type Subsystem interface {
	Start() error
	Stop() error
}

type ShowManager struct {
	bus        *server.Server
	location   string
	subsystems []Subsystem
}

func NewShowManager(location string) (*ShowManager, error) {

	// Create the directory if it doesn't exist
	about, err := os.Stat(location)
	if errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(location, 0755)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	} else if !about.IsDir() {
		return nil, errors.New("location is not a directory")
	}

	comBus, err := bus.NewBus()
	if err != nil {
		return nil, err
	}

	manager := &ShowManager{
		bus:      comBus,
		location: location,
	}

	err = errors.Join(
		manager.addSubsystem(func(location string, dbConn *nats.Conn) (Subsystem, error) {
			return database.NewDatabase(location, dbConn)
		}),
	)

	if err != nil {
		slog.Error("failed to start subsystems", "error", err)
		return nil, err
	}

	return manager, err
}

func (m *ShowManager) addSubsystem(fn func(string, *nats.Conn) (Subsystem, error)) error {
	dbConn, err := bus.GetInProcessConn(m.bus)
	if err != nil {
		return err
	}

	subsystem, err := fn(m.location, dbConn)
	if err != nil {
		return err
	}
	m.subsystems = append(m.subsystems, subsystem)
	err = subsystem.Start()
	if err != nil {
		return err
	}

	return nil
}

func (m *ShowManager) GetBusConnection() (*nats.Conn, error) {
	return bus.GetInProcessConn(m.bus)
}

func (m *ShowManager) Stop() error {
	m.bus.Shutdown()
	for _, s := range m.subsystems {
		s.Stop()
	}
	return nil
}
