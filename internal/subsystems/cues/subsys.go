package cues

import (
	"errors"

	"github.com/nats-io/nats.go"
	"github.com/tidwall/buntdb"
	"gitlab.com/stexxo/dynocue/internal/data"
)

const (
	dbSchemaVersion = "1.0"
)

type CueSubsystem struct{}

func NewCueSubsystem() *CueSubsystem {
	return &CueSubsystem{}
}

func (c *CueSubsystem) Start(conn *nats.Conn, db *buntdb.DB, savePath string) error {
	return nil
}

func (c *CueSubsystem) setupDb(db *buntdb.DB) error {
	// Check for Setup Indicator
	val, found, err := data.GetConfiguredVersion(db, "cues")
	if err != nil {
		return err
	}
	if found && val == dbSchemaVersion {
		return nil
	}

	// This is where a migration would be determined to be needed
	// 1.0 for now so a migration is not required
	return c.dbFreshConfigure(db)
}

func (c *CueSubsystem) dbFreshConfigure(db *buntdb.DB) error {
	err := errors.Join(
		db.CreateIndex("cue_numbers", "cuelist:1:cues:*", buntdb.IndexJSON("number")),
	)
}

func (c *CueSubsystem) Stop() error {
	return nil
}
