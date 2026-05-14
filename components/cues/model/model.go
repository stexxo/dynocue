// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package model

import (
	"sync"

	"github.com/hashicorp/go-memdb"
	"github.com/stexxo/dynocue/util"
)

type CueingModel struct {
	dbMu       *sync.RWMutex // All R&W on the models should use RLock. Write Lock should be used for large db operations such preventing Reads or Writes during Saving/loading
	persistent *memdb.MemDB
	runtime    *memdb.MemDB
	registry   *util.EventRegistry
}

func NewCueingModel() (*CueingModel, error) {
	pdb, err := memdb.NewMemDB(persistentSchema)
	if err != nil {
		return nil, err
	}

	rdb, err := memdb.NewMemDB(runtimeSchema)
	if err != nil {
		return nil, err
	}

	return &CueingModel{
		dbMu:       &sync.RWMutex{},
		persistent: pdb,
		runtime:    rdb,
		registry:   util.NewEventRegistry(),
	}, nil
}
