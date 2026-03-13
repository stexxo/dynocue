package cues

import (
	"github.com/nats-io/nats.go"
	"go.etcd.io/bbolt"
)

type Cues struct {
	conn *nats.Conn
	db   *bbolt.DB
}

func NewCues(conn *nats.Conn, db *bbolt.DB) (*Cues, error) {

	return &Cues{
		conn: conn,
		db:   db,
	}, nil
}

func (c Cues) Close() error {
	return nil
}
