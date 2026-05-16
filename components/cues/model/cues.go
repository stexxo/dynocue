// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package model

import (
	"errors"

	"github.com/google/uuid"
	"github.com/hashicorp/go-memdb"
	"github.com/stexxo/dynocue/components/cues/types"
	"github.com/stexxo/dynocue/db"
)

var ErrCueNotFound = errors.New("cue not found")

func (m *CueingModel) CreateCue(cueListId string, number uint) (string, uint, error) {
	m.dbMu.RLock()
	defer m.dbMu.RUnlock()
	cue := types.Cue{
		CueListId: cueListId,
		CueId:     uuid.NewString(),
	}

	err := db.WithWrite(m.persistent, func(txn *memdb.Txn) error {
		// Check that CueList Exists
		item, err := txn.First(TableCueLists, IndexId, cueListId)
		if err != nil {
			return err
		}
		if item == nil {
			return ErrCueListNotFound
		}

		num, err := getNextNumber[types.Cue](txn, number, TableCues, IndexNumberPrefix, []any{cueListId}, IndexNumber, []any{cueListId, number}, func(t *types.Cue) uint {
			return t.Number
		})
		if err != nil {
			return err
		}
		cue.Number = num
		if err := txn.Insert(TableCues, &cue); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return "", 0, err
	}

	m.registry.Emit(ResourceCue, OperationCreated, MetadataCueListId, cue.CueListId, MetadataCueId, cue.CueId)
	return cue.CueId, cue.Number, nil
}

func (m *CueingModel) EnumerateCues(cueListId string) ([]types.Cue, error) {
	m.dbMu.RLock()
	defer m.dbMu.RUnlock()
	return db.GetAllDb[types.Cue](m.persistent, TableCues, IndexNumberPrefix, cueListId)
}

func (m *CueingModel) GetCueByNumber(cueListId string, number uint) (*types.Cue, error) {
	m.dbMu.RLock()
	defer m.dbMu.RUnlock()
	out, err := db.GetFirstDb[types.Cue](m.persistent, TableCues, IndexNumber, cueListId, number)
	if errors.Is(err, db.ErrItemNotFound) {
		return nil, ErrCueNotFound
	}
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (m *CueingModel) GetCueById(cueId string) (*types.Cue, error) {
	m.dbMu.RLock()
	defer m.dbMu.RUnlock()
	out, err := db.GetFirstDb[types.Cue](m.persistent, TableCues, IndexId, cueId)
	if errors.Is(err, db.ErrItemNotFound) {
		return nil, ErrCueNotFound
	}
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (m *CueingModel) GetFirstCueInCueList(cueListId string) (*types.Cue, error) {
	m.dbMu.RLock()
	defer m.dbMu.RUnlock()
	out, err := db.WithRead[*types.Cue](m.persistent, func(txn *memdb.Txn) (*types.Cue, error) {
		// Check that CueList exists
		_, err := db.GetFirstTxn[types.CueList](txn, TableCueLists, IndexId, cueListId)
		if errors.Is(err, db.ErrItemNotFound) {
			return nil, ErrCueListNotFound
		}
		if err != nil {
			return nil, err
		}

		// Get The First Cue
		res, err := db.GetFirstTxn[types.Cue](txn, TableCues, IndexNumberPrefix, cueListId)
		if errors.Is(err, db.ErrItemNotFound) {
			return nil, ErrCueNotFound
		}
		if err != nil {
			return nil, err
		}

		return res, nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

var ErrNoNextCue = errors.New("no next cue found")

func (m *CueingModel) GetNextCueInCueList(cueListId string, cueId string) (*types.Cue, error) {
	m.dbMu.RLock()
	defer m.dbMu.RUnlock()
	if cueId == "" {
		return m.GetFirstCueInCueList(cueListId)
	}

	out, err := db.WithRead[*types.Cue](m.persistent, func(txn *memdb.Txn) (*types.Cue, error) {
		// Get Cue List and Make Sure it Exists
		cl, err := db.GetFirstTxn[types.CueList](txn, TableCueLists, IndexId, cueListId)
		if errors.Is(err, db.ErrItemNotFound) {
			return nil, ErrCueListNotFound
		}
		if err != nil {
			return nil, err
		}

		// Get the current cue to retrieve the number from it
		current, err := db.GetFirstTxn[types.Cue](txn, TableCues, IndexId, cueId)
		if errors.Is(err, db.ErrItemNotFound) {
			return nil, ErrCueNotFound
		}
		if err != nil {
			return nil, err
		}

		// Start at the lower bound being the current number
		iter, err := txn.LowerBound(TableCues, IndexNumber, cueListId, current.Number)
		if err != nil {
			return nil, err
		}
		iter.Next() // skip current
		out := iter.Next()

		// If there is no next cue handle appropriately
		if out == nil {

			// If the Cuelist is configured to wrap, then go get the first cue in the list
			if cl.WrapAtEnd {
				res, err := db.GetFirstTxn[types.Cue](txn, TableCues, IndexNumberPrefix, cueListId)
				if errors.Is(err, db.ErrItemNotFound) {
					return nil, ErrCueNotFound
				}
				if err != nil {
					return nil, err
				}
				return res, nil
			}

			// if it is not configured to wrap, then return no next cue error
			return nil, ErrNoNextCue
		}

		// Return the cue from the iteration if it is not nil
		return out.(*types.Cue), nil
	})

	if err != nil {
		return nil, err
	}

	return out, nil
}

func (m *CueingModel) DeleteCueById(cueId string) error {
	m.dbMu.RLock()
	defer m.dbMu.RUnlock()
	cue, err := m.GetCueById(cueId)
	if errors.Is(err, ErrCueNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	err = db.DeleteItemFromDb[types.Cue](m.persistent, TableCues, IndexId, cueId)
	if err != nil {
		return err
	}
	m.registry.Emit(ResourceCue, OperationDeleted, MetadataCueListId, cue.CueListId, MetadataCueId, cueId)
	return nil
}

func (m *CueingModel) DeleteAllCuesByCueListId(cueListId string) error {
	m.dbMu.RLock()
	defer m.dbMu.RUnlock()
	deleted := map[string][]string{}

	err := db.WithWrite(m.persistent, func(txn *memdb.Txn) error {

		// Delete All Cues
		cues, err := db.GetAllTxn[types.Cue](txn, TableCues, IndexNumberPrefix, cueListId)
		if err != nil {
			return err
		}
		for _, cue := range cues {
			deleted[cue.CueId] = []string{}
			err := db.DeleteItemFromTxn[types.Cue](txn, TableCues, IndexId, cue.CueId)
			if err != nil {
				return err
			}

			// Delete All Actions
			actions, err := db.GetAllTxn[types.Action](txn, TableActions, IndexNumberPrefix, cue.CueId)
			if err != nil {
				return err
			}
			for _, action := range actions {
				deleted[cue.CueId] = append(deleted[cue.CueId], action.ActionId)
				err := db.DeleteItemFromTxn[types.Action](txn, TableActions, IndexId, action.ActionId)
				if err != nil {
					return err
				}
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	for cueId, actions := range deleted {
		m.registry.Emit(ResourceCue, OperationDeleted, MetadataCueListId, cueListId, MetadataCueId, cueId)
		for _, actionid := range actions {
			m.registry.Emit(ResourceAction, OperationDeleted, MetadataCueListId, cueListId, MetadataCueId, cueId, MetadataActionId, actionid)
		}
	}

	return nil
}

func (m *CueingModel) UpdateCueAttribute(cueId string, field string, value interface{}) error {
	m.dbMu.RLock()
	defer m.dbMu.RUnlock()
	c, err := db.UpdateStructDb[types.Cue](m.persistent, TableCues, IndexId, cueId, field, value)
	if errors.Is(err, db.ErrItemNotFound) {
		return ErrCueNotFound
	}
	if err != nil {
		return err
	}

	m.registry.Emit(ResourceCue, OperationUpdated, MetadataCueListId, c.CueListId, MetadataCueId, cueId)
	return nil
}
