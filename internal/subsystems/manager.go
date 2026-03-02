package subsystems

import (
	"errors"
	"os"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"github.com/tidwall/buntdb"
	"gitlab.com/stexxo/dynocue/internal/bus"
	"gitlab.com/stexxo/dynocue/internal/subsystems/cues"
)

// Subsystem is an interface that all subsystems must implement
type Subsystem interface {
	Start(*nats.Conn, *buntdb.DB, string) error
	Stop() error
}

// ShowManager manages the lifecycle of a show
type ShowManager struct {
	bus        *server.Server
	location   string
	db         *buntdb.DB
	subsystems []Subsystem
}

// NewShowManager creates a new show manager for a provided save path
func NewShowManager(savePath string) (*ShowManager, error) {
	// Create the directory if it doesn't exist
	about, err := os.Stat(savePath)
	if errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(savePath, 0755)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	} else if !about.IsDir() {
		return nil, errors.New("savePath is not a directory")
	}

	// Build Communication Bus For the Show
	comBus, err := bus.NewBus()
	if err != nil {
		return nil, err
	}

	// Open Database
	db, err := buntdb.Open(savePath + "/dynocue.db")
	if err != nil {
		return nil, err
	}
	err = db.SetConfig(buntdb.Config{
		SyncPolicy:           buntdb.Always,
		AutoShrinkPercentage: 100,
		AutoShrinkMinSize:    32 * 1024 * 1024,
		AutoShrinkDisabled:   false,
	})
	if err != nil {
		return nil, err
	}

	// Build Show Manager
	manager := &ShowManager{
		bus:      comBus,
		location: savePath,
		db:       db,
		subsystems: []Subsystem{
			cues.NewCueSubsystem(),
		},
	}

	// Start Subsystems managed by Show Manager
	for _, s := range manager.subsystems {
		conn, err := manager.GetBusConnection()
		if err != nil {
			return nil, err
		}

		err = s.Start(conn, manager.db, manager.location)
		if err != nil {
			return nil, err
		}
	}

	return manager, err
}

// GetBusConnection is a helper function to get a connection to the communication bus
func (m *ShowManager) GetBusConnection() (*nats.Conn, error) {
	return bus.GetInProcessConn(m.bus)
}

// Stop shuts down the show manager. It cannot be restarted
func (m *ShowManager) Stop() error {
	m.bus.Shutdown()
	m.db.Close()

	var err error
	for _, s := range m.subsystems {
		err = errors.Join(err, s.Stop())
	}

	return err
}
