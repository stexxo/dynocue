package model

import (
	"errors"
	"io"

	"github.com/hashicorp/go-memdb"
	"github.com/stexxo/dynocue/components/cues/types"
	"github.com/stexxo/dynocue/db"
)

func (m *CueingModel) SerializeEachTable(fn func(name string, reader io.Reader) error) error {
	for _, table := range persistentSchema.Tables {
		buf, err := db.SerializeTable(m.persistent, table.Name)
		if err != nil {
			return err
		}
		err = fn(table.Name, buf)
		if err != nil {
			return err
		}
	}
	return nil
}

type tableRestorer func(memDb *memdb.MemDB, reader io.Reader) error

var tableRestore = map[string]tableRestorer{
	TableCueLists: func(memDb *memdb.MemDB, reader io.Reader) error {
		return db.RestoreTable[types.CueList](memDb, TableCueLists, reader)
	},
	TableCues: func(memDb *memdb.MemDB, reader io.Reader) error {
		return db.RestoreTable[types.Cue](memDb, TableCues, reader)
	},
	TableActions: func(memDb *memdb.MemDB, reader io.Reader) error {
		return db.RestoreTable[types.Action](memDb, TableActions, reader)
	},
}

func (m *CueingModel) RestoreTable(name string, data io.Reader) error {
	fn, ok := tableRestore[name]
	if !ok {
		return errors.New("table not found")
	}

	err := fn(m.persistent, data)
	if err != nil {
		return err
	}
	return nil
}
