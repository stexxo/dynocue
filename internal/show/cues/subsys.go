package cues

import (
	"github.com/nats-io/nats.go"
	apibus "gitlab.com/stexxo/dynocue/api/bus"
	apicues "gitlab.com/stexxo/dynocue/api/cues"
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

	if _, err := apibus.Reply(conn, apicues.RequestCreateCueList, c.NewCueList); err != nil {
		return nil, err
	}
	if _, err := apibus.Reply(conn, apicues.RequestUpdateCueListMetadata, c.UpdateCueListMetadata); err != nil {
		return nil, err
	}
	if _, err := apibus.Reply(conn, apicues.RequestGetCueListMetadata, c.GetCueListMetadata); err != nil {
		return nil, err
	}
	if _, err := apibus.Reply(conn, apicues.RequestEnumerateCueList, c.EnumerateCueList); err != nil {
		return nil, err
	}
	if _, err := apibus.Reply(conn, apicues.RequestDeleteCueList, c.DeleteCueList); err != nil {
		return nil, err
	}
	if _, err := apibus.Reply(conn, apicues.RequestMoveCueList, c.MoveCueList); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *CueSystem) Close() error {
	return nil
}
