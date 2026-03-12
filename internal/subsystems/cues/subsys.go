package cues

import (
	"errors"

	"github.com/nats-io/nats.go"
	"go.etcd.io/bbolt"
)

type CueSubsystem struct {
	started  bool
	conn     *nats.Conn
	db       *bbolt.DB
	savePath string
}

func NewCueSubsystem() *CueSubsystem {
	return &CueSubsystem{}
}

func (c *CueSubsystem) Start(conn *nats.Conn, db *bbolt.DB, savePath string) error {
	if c.started {
		return errors.New("subsystem already started")
	}

	err := db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("cuelists"))
		return err
	})
	if err != nil {
		return err
	}

	c.conn = conn
	c.db = db
	c.savePath = savePath
	c.started = true

	_, err = NewCueListHandlers(conn, db)
	return err
}

func (c *CueSubsystem) Stop() error {
	return nil
}
