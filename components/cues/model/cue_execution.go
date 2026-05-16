// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package model

import (
	"errors"
	"time"

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

		// Check if it is already running, and if so, return nil, idempotent start
		existingExecution, err := db.GetFirstDb[types.CueExecution](m.persistent, TableActionExecution, IndexId, cueId)
		if err != nil && !errors.Is(err, db.ErrItemNotFound) {
			return err
		}
		if existingExecution != nil {
			return nil
		}

		// Insert the Cue Execution
		ce := &types.CueExecution{
			CueListId:    cue.CueListId,
			CueId:        cueId,
			Selected:     selected,
			Active:       active,
			CueExecStart: time.Now(),
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
						replacement := &types.CueExecution{
							CueListId:    selectedCue.CueListId,
							CueId:        selectedCue.CueId,
							Active:       selectedCue.Active,
							Selected:     false,
							CueExecStart: selectedCue.CueExecStart,
							DelayActive:  selectedCue.DelayActive,
							DelayStart:   selectedCue.DelayStart,
							FollowActive: selectedCue.FollowActive,
							FollowStart:  selectedCue.FollowStart,
						}
						err = tx.Insert(TableCueExecution, replacement)
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
	if errors.Is(err, db.ErrItemNotFound) {
		return nil, ErrCueNotFound
	}
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

func (m *CueingModel) StartCueDelayExecution(cueId string) error {
	var updatedCueExec *types.CueExecution
	err := db.WithWrite(m.runtime, func(tx *memdb.Txn) error {
		cueExec, err := db.GetFirstTxn[types.CueExecution](tx, TableCueExecution, IndexId, cueId)
		if errors.Is(err, db.ErrItemNotFound) {
			return ErrCueNotFound
		}
		if err != nil {
			return err
		}

		updatedCueExec = &types.CueExecution{
			CueListId:    cueExec.CueListId,
			CueId:        cueExec.CueId,
			Active:       cueExec.Active,
			Selected:     cueExec.Selected,
			CueExecStart: cueExec.CueExecStart,
			DelayActive:  true,
			DelayStart:   time.Now(),
			FollowActive: cueExec.FollowActive,
			FollowStart:  cueExec.FollowStart,
		}
		return tx.Insert(TableCueExecution, updatedCueExec)
	})

	// emit updated event
	m.registry.Emit(ResourceCueExecution, OperationUpdated, MetadataCueListId, updatedCueExec.CueListId, MetadataCueId, updatedCueExec.CueId)

	return err
}

func (m *CueingModel) StopCueDelayExecution(cueId string) error {
	var updatedCueExec *types.CueExecution
	err := db.WithWrite(m.runtime, func(tx *memdb.Txn) error {
		cueExec, err := db.GetFirstTxn[types.CueExecution](tx, TableCueExecution, IndexId, cueId)
		if errors.Is(err, db.ErrItemNotFound) {
			return ErrCueNotFound
		}
		if err != nil {
			return err
		}

		updatedCueExec = &types.CueExecution{
			CueListId:    cueExec.CueListId,
			CueId:        cueExec.CueId,
			Active:       cueExec.Active,
			Selected:     cueExec.Selected,
			CueExecStart: cueExec.CueExecStart,
			DelayActive:  false,
			DelayStart:   cueExec.DelayStart,
			FollowActive: cueExec.FollowActive,
			FollowStart:  cueExec.FollowStart,
		}
		return tx.Insert(TableCueExecution, updatedCueExec)
	})

	// emit updated event
	m.registry.Emit(ResourceCueExecution, OperationUpdated, MetadataCueListId, updatedCueExec.CueListId, MetadataCueId, updatedCueExec.CueId)

	return err
}

func (m *CueingModel) StartCueFollowExecution(cueId string, delay time.Duration) error {
	var updatedCueExec *types.CueExecution
	err := db.WithWrite(m.runtime, func(tx *memdb.Txn) error {
		cueExec, err := db.GetFirstTxn[types.CueExecution](tx, TableCueExecution, IndexId, cueId)
		if errors.Is(err, db.ErrItemNotFound) {
			return ErrCueNotFound
		}
		if err != nil {
			return err
		}

		updatedCueExec = &types.CueExecution{
			CueListId:    cueExec.CueListId,
			CueId:        cueExec.CueId,
			Active:       cueExec.Active,
			Selected:     cueExec.Selected,
			CueExecStart: cueExec.CueExecStart,
			DelayActive:  cueExec.DelayActive,
			DelayStart:   cueExec.DelayStart,
			FollowActive: true,
			FollowStart:  time.Now(),
		}
		return tx.Insert(TableCueExecution, updatedCueExec)
	})

	// emit updated event
	m.registry.Emit(ResourceCueExecution, OperationUpdated, MetadataCueListId, updatedCueExec.CueListId, MetadataCueId, updatedCueExec.CueId)

	return err
}

func (m *CueingModel) StopCueFollowExecution(cueId string) error {
	var updatedCueExec *types.CueExecution
	err := db.WithWrite(m.runtime, func(tx *memdb.Txn) error {
		cueExec, err := db.GetFirstTxn[types.CueExecution](tx, TableCueExecution, IndexId, cueId)
		if errors.Is(err, db.ErrItemNotFound) {
			return ErrCueNotFound
		}
		if err != nil {
			return err
		}

		updatedCueExec = &types.CueExecution{
			CueListId:    cueExec.CueListId,
			CueId:        cueExec.CueId,
			Active:       cueExec.Active,
			Selected:     cueExec.Selected,
			CueExecStart: cueExec.CueExecStart,
			DelayActive:  cueExec.DelayActive,
			DelayStart:   cueExec.DelayStart,
			FollowActive: false,
			FollowStart:  cueExec.FollowStart,
		}

		return tx.Insert(TableCueExecution, updatedCueExec)
	})

	// emit updated event
	m.registry.Emit(ResourceCueExecution, OperationUpdated, MetadataCueListId, updatedCueExec.CueListId, MetadataCueId, updatedCueExec.CueId)

	return err
}
