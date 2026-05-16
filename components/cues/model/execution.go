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
	err := db.WithWrite(m.persistent, func(tx *memdb.Txn) error {
		// Check that Cue Exists
		cue, err := db.GetFirstTxn[types.Cue](tx, TableCues, IndexId, cueId)
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
						err = tx.Delete(TableCueExecution, selectedCue.CueId)
						if err != nil {
							return err
						}
						executionsRemoved = append(executionsRemoved, selectedCue.CueId)
					} else { // if the cue is active, and now not selected, clear the selection but leave the execution record
						_, err := db.UpdateStructTxn[types.CueExecution](tx, TableCueExecution, IndexId, selectedCue.CueId, "Selected", false)
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

func (m *CueingModel) GetSelectedCue(cueListId string) (string, error) {
	res, err := db.GetFirstDb[types.CueExecution](m.persistent, TableCueExecution, IndexSelected, cueListId, true)
	if err != nil {
		return "", err
	}
	return res.CueId, nil
}

func (m *CueingModel) StopCueExecution(cueId string) error {
	var deleted bool
	var cueListId string
	err := db.WithWrite(m.persistent, func(tx *memdb.Txn) error {
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
			return tx.Delete(TableCueExecution, cueId)
		}

		// If it is still selected, make a copy and insert the finished state into db
		newCueExecution := &types.CueExecution{
			CueListId: cueExecution.CueListId,
			CueId:     cueExecution.CueId,
			Selected:  cueExecution.Selected,
		}

		return tx.Insert(TableCueExecution, &newCueExecution)
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
