package model

import (
	"sync"

	"github.com/hashicorp/go-memdb"
	"github.com/stexxo/dynocue/util"
)

type AudioModel struct {
	dbMu       *sync.RWMutex
	persistent *memdb.MemDB
	runtime    *memdb.MemDB
	registry   *util.EventRegistry
}

func NewAudioModel() (*AudioModel, error) {
	pdb, err := memdb.NewMemDB(persistentSchema)
	if err != nil {
		return nil, err
	}
	rdb, err := memdb.NewMemDB(runtimeSchema)
	if err != nil {
		return nil, err
	}
	am := &AudioModel{
		registry:   util.NewEventRegistry(),
		dbMu:       &sync.RWMutex{},
		persistent: pdb,
		runtime:    rdb,
	}
	return am, nil
}
