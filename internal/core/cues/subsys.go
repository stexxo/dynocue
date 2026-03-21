// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package cues

import (
	"github.com/nats-io/nats.go"
	apicues "github.com/stexxo/dynocue/api/cues"
	apibus "github.com/stexxo/dynocue/pkg/bus"
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
	if _, err := apibus.Reply(conn, apicues.RequestUpdateCueList, c.UpdateCueList); err != nil {
		return nil, err
	}
	if _, err := apibus.Reply(conn, apicues.RequestGetCueList, c.GetCueList); err != nil {
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
	if _, err := apibus.Reply(conn, apicues.RequestUpdateCue, c.UpdateCue); err != nil {
		return nil, err
	}
	if _, err := apibus.Reply(conn, apicues.RequestGetCue, c.GetCue); err != nil {
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
	if _, err := apibus.Reply(conn, apicues.RequestCreateAction, c.NewAction); err != nil {
		return nil, err
	}
	if _, err := apibus.Reply(conn, apicues.RequestUpdateAction, c.UpdateAction); err != nil {
		return nil, err
	}
	if _, err := apibus.Reply(conn, apicues.RequestGetAction, c.GetAction); err != nil {
		return nil, err
	}
	if _, err := apibus.Reply(conn, apicues.RequestEnumerateAction, c.EnumerateAction); err != nil {
		return nil, err
	}
	if _, err := apibus.Reply(conn, apicues.RequestDeleteAction, c.DeleteAction); err != nil {
		return nil, err
	}
	if _, err := apibus.Reply(conn, apicues.RequestMoveAction, c.MoveAction); err != nil {
		return nil, err
	}

	return c, nil
}

// Close performs any necessary cleanup for the CueSystem.
func (c *CueSystem) Close() error {
	return nil
}
