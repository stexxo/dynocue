package model

import "github.com/hashicorp/go-memdb"

type CueingModel struct {
	persistent *memdb.MemDB
	runtime    *memdb.MemDB
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
		persistent: pdb,
		runtime:    rdb,
	}, nil
}
