package cues

import (
	"github.com/nats-io/nats.go"
	apicues "gitlab.com/stexxo/dynocue/api/cues"
	"gitlab.com/stexxo/dynocue/internal/bus"
	"go.etcd.io/bbolt"
)

type CueSystem struct {
	conn *nats.Conn
	db   *bbolt.DB
}

func NewCues(conn *nats.Conn, db *bbolt.DB) (*CueSystem, error) {
	c := &CueSystem{
		conn: conn,
		db:   db,
	}

	if _, err := bus.Reply(conn, apicues.RequestCreateCueList, c.NewCueList); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *CueSystem) Close() error {
	return nil
}
