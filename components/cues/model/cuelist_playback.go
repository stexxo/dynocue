package model

import (
	"errors"

	"github.com/hashicorp/go-memdb"
	"github.com/stexxo/dynocue/components/cues/types"
	"github.com/stexxo/dynocue/db"
)

func (m *CueingModel) SetSelectedCueId(cueListId string, selectedCueId string) error {
	err := db.WithWrite(m.runtime, func(txn *memdb.Txn) error {
		clp := &types.CueListSelectedCue{
			CueListId:     cueListId,
			SelectedCueId: selectedCueId,
		}
		err := txn.Insert(TableCueListPlayback, clp)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return err
	}

	m.registry.Emit(ResourceCueListCueSelection, OperationUpdated, MetadataCueListId, cueListId, MetadataCueId, selectedCueId)
	return nil
}

func (m *CueingModel) GetSelectedCueId(cueListId string) (*types.CueListSelectedCue, error) {
	item, err := db.GetFirstDb[types.CueListSelectedCue](m.runtime, TableCueListPlayback, IndexId, cueListId)
	if errors.Is(err, db.ErrItemNotFound) {
		return &types.CueListSelectedCue{
			CueListId: cueListId,
		}, nil
	}
	if err != nil {
		return nil, err
	}

	return item, err
}

func (m *CueingModel) ClearSelectedCueId(cueListId string) error {
	err := db.DeleteItemFromDb[*types.CueListSelectedCue](m.runtime, TableCueListPlayback, IndexId, cueListId)
	if err != nil {
		return err
	}
	m.registry.Emit(ResourceCueListCueSelection, OperationDeleted, MetadataCueListId, cueListId)
	return nil
}
