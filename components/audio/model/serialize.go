// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package model

import (
	"errors"
	"io"

	"github.com/hashicorp/go-memdb"
	"github.com/stexxo/dynocue/components/audio/types"
	"github.com/stexxo/dynocue/db"
)

func (m *AudioModel) SerializeEachTable(fn func(name string, reader io.Reader) error) error {
	m.dbMu.Lock()
	defer m.dbMu.Unlock()

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
	TableFiles: func(memDb *memdb.MemDB, reader io.Reader) error {
		return db.RestoreTable[types.AudioFile](memDb, TableFiles, reader)
	},
}

func (m *AudioModel) RestoreTable(name string, data io.Reader) error {
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

func (m *AudioModel) LoadModel(fn func(name string) (io.Reader, error)) error {
	// Clear the Table
	m.dbMu.Lock()
	defer m.dbMu.Unlock()

	var err error
	m.persistent, err = memdb.NewMemDB(persistentSchema)
	if err != nil {
		return err
	}

	m.runtime, err = memdb.NewMemDB(runtimeSchema)
	if err != nil {
		return err
	}

	var errs error
	for tableName, restorer := range tableRestore {
		res, err := fn(tableName)
		if err != nil {
			errs = errors.Join(errs, err)
			continue
		}
		err = restorer(m.persistent, res)
		if err != nil {
			errs = errors.Join(errs, err)
		}
	}
	m.registry.Emit(ResourceModel, OperationLoaded)
	return errs
}
