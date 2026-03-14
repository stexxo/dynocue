package cues

import (
	"github.com/nats-io/nats.go"
	apibus "gitlab.com/stexxo/dynocue/api/bus"
	apicues "gitlab.com/stexxo/dynocue/api/cues"
	"go.etcd.io/bbolt"
)

// CueSystem manages the cue and cue list operations, handling communication
// between the NATS messaging bus and the bbolt database.
type CueSystem struct {
	conn *nats.Conn
	db   *bbolt.DB
}

// NewCues initializes a new CueSystem, setting up NATS responders for all
// cue and cue list related requests.
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
	if _, err := apibus.Reply(conn, apicues.RequestCreateCue, c.NewCue); err != nil {
		return nil, err
	}
	if _, err := apibus.Reply(conn, apicues.RequestUpdateCueMetadata, c.UpdateCueMetadata); err != nil {
		return nil, err
	}
	if _, err := apibus.Reply(conn, apicues.RequestGetCueMetadata, c.GetCueMetadata); err != nil {
		return nil, err
	}
	if _, err := apibus.Reply(conn, apicues.RequestEnumerateCue, c.EnumerateCue); err != nil {
		return nil, err
	}
	if _, err := apibus.Reply(conn, apicues.RequestDeleteCue, c.DeleteCue); err != nil {
		return nil, err
	}
	if _, err := apibus.Reply(conn, apicues.RequestMoveCue, c.MoveCue); err != nil {
		return nil, err
	}

	return c, nil
}

// Close performs any necessary cleanup for the CueSystem.
func (c *CueSystem) Close() error {
	return nil
}
