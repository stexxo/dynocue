package show

import (
	"errors"
	"fmt"
	"os"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	ibus "gitlab.com/stexxo/dynocue/internal/bus"
	"gitlab.com/stexxo/dynocue/internal/show/cues"
	"go.etcd.io/bbolt"
)

type Subsystem interface {
	Close() error
}

type Show struct {
	db       *bbolt.DB
	bus      *server.Server
	savePath string

	subsystem []Subsystem
}

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

func (s *Show) GetConn() (*nats.Conn, error) {
	return ibus.GetInProcessConn(s.bus)
}

func (s *Show) Close() error {
	for _, subsystem := range s.subsystem {
		if err := subsystem.Close(); err != nil {
			return err
		}
	}

	s.bus.Shutdown()
	return s.db.Close()
}
