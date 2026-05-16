// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package model

import (
	"errors"

	"github.com/hashicorp/go-memdb"
	"github.com/stexxo/dynocue/components/cues/types"
	"github.com/stexxo/dynocue/db"
)

func (m *CueingModel) StartCueExecution(cueId string, selected bool, active bool) error {
	var cueListId string
	var executionsRemoved []string
	var executionsUnselected []string
	err := db.WithWrite(m.runtime, func(tx *memdb.Txn) error {
		// Check that Cue Exists
		cue, err := db.GetFirstDb[types.Cue](m.persistent, TableCues, IndexId, cueId)
		if errors.Is(err, db.ErrItemNotFound) {
			return ErrCueNotFound
		}
		if err != nil {
			return err
		}
		cueListId = cue.CueListId

		// Insert the Cue Execution
		ce := &types.CueExecution{
			CueListId: cue.CueListId,
			CueId:     cueId,
			Selected:  selected,
			Active:    active,
			Elapsed:   0,
		}
		err = tx.Insert(TableCueExecution, ce)
		if err != nil {
			return err
		}

		// If this cue is to be selected, look for other cues that are also selected and clear their selection
		if selected {
			res, err := db.GetAllTxn[types.CueExecution](tx, TableCueExecution, IndexSelected, cue.CueListId, true)
			if err != nil {
				return err
			}

			for _, selectedCue := range res {
				if selectedCue.CueId != cueId { // ensure what is found is not what we just set
					if !selectedCue.Active { // if the cue is not active and now not selected, delete the execution record
						err = db.DeleteItemFromTxn[types.CueExecution](tx, TableCueExecution, IndexId, selectedCue.CueId)
						if err != nil {
							return err
						}
						executionsRemoved = append(executionsRemoved, selectedCue.CueId)
					} else { // if the cue is active, and now not selected, clear the selection but leave the execution record
						_, err := db.UpdateStructTxn[types.CueExecution](tx, TableCueExecution, IndexId, selectedCue.CueId, "selected", false)
						if err != nil {
							return err
						}
						executionsUnselected = append(executionsUnselected, selectedCue.CueId)
					}
				}
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	m.registry.Emit(ResourceCueExecution, OperationStarted, MetadataCueListId, cueListId, MetadataCueId, cueId)
	for _, executionId := range executionsRemoved {
		m.registry.Emit(ResourceCueExecution, OperationDeleted, MetadataCueListId, cueListId, MetadataCueId, executionId)
	}
	for _, executionId := range executionsUnselected {
		m.registry.Emit(ResourceCueExecution, OperationUnselected, MetadataCueListId, cueListId, MetadataCueId, executionId)
	}
	return nil
}

func (m *CueingModel) GetSelectedCue(cueListId string) (*types.CueExecution, error) {
	res, err := db.GetFirstDb[types.CueExecution](m.runtime, TableCueExecution, IndexSelected, cueListId, true)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (m *CueingModel) GetCueExecution(cueId string) (*types.CueExecution, error) {
	res, err := db.GetFirstDb[types.CueExecution](m.runtime, TableCueExecution, IndexId, cueId)
	if errors.Is(err, db.ErrItemNotFound) {
		return nil, ErrCueNotFound
	}
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (m *CueingModel) EnumerateCueExecutions(cueListId string) ([]types.CueExecution, error) {
	res, err := db.GetAllDb[types.CueExecution](m.runtime, TableCueExecution, IndexCueList, cueListId)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (m *CueingModel) StopCueExecution(cueId string) error {
	var deleted bool
	var cueListId string
	err := db.WithWrite(m.runtime, func(tx *memdb.Txn) error {
		// Check that the cue execution exists
		cueExecution, err := db.GetFirstTxn[types.CueExecution](tx, TableCueExecution, IndexId, cueId)
		if errors.Is(err, db.ErrItemNotFound) {
			return nil // idempotent - already finished if it is deleted
		}
		if err != nil {
			return err
		}
		cueListId = cueExecution.CueListId

		if !cueExecution.Selected { // if the cue is not selected, and now is not active, delete the execution record
			deleted = true
			return db.DeleteItemFromTxn[types.CueExecution](tx, TableCueExecution, IndexId, cueExecution.CueId)
		}

		// If it is still selected, make a copy and insert the finished state into db
		newCueExecution := &types.CueExecution{
			CueListId: cueExecution.CueListId,
			CueId:     cueExecution.CueId,
			Selected:  cueExecution.Selected,
		}

		return tx.Insert(TableCueExecution, newCueExecution)
	})
	if err != nil {
		return err
	}

	if deleted {
		m.registry.Emit(ResourceCueExecution, OperationDeleted, MetadataCueListId, cueListId, MetadataCueId, cueId)
	} else {
		m.registry.Emit(ResourceCueExecution, OperationFinished, MetadataCueListId, cueListId, MetadataCueId, cueId)
	}

	return nil
}
