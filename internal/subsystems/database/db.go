package database

import (
	"path"

	"github.com/dgraph-io/badger/v4"
	"github.com/nats-io/nats.go"
)

type Database struct {
	db    *badger.DB
	evBus *nats.Conn
}

func NewDatabase(showPath string, evBus *nats.Conn) (*Database, error) {
	db, err := badger.Open(badger.DefaultOptions(path.Join(showPath, "db")).WithBypassLockGuard(true))
	if err != nil {
		return nil, err
	}

	return &Database{
		db:    db,
		evBus: evBus,
	}, nil
}

func (d *Database) Start() error {
	return nil
}

func (d *Database) Stop() error {
	return d.db.Close()
}
