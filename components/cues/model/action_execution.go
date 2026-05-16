package model

import (
	"errors"
	"time"

	"github.com/hashicorp/go-memdb"
	"github.com/stexxo/dynocue/components/cues/types"
	"github.com/stexxo/dynocue/db"
)

func (m *CueingModel) StartActionExecution(actionId string) error {
	cueListId := ""
	cueId := ""
	err := db.WithWrite(m.runtime, func(tx *memdb.Txn) error {
		// Check that Action Exists
		action, err := db.GetFirstDb[types.Action](m.persistent, TableActions, IndexId, actionId)
		if err != nil {
			return err
		}
		cueListId = action.CueListId
		cueId = action.CueId

		// Check if it is already running, and if so, return nil, idempotent start
		existingExecution, err := db.GetFirstDb[types.ActionExecution](m.persistent, TableActionExecution, IndexId, actionId)
		if err != nil && !errors.Is(err, db.ErrItemNotFound) {
			return err
		}
		if existingExecution != nil {
			return nil
		}

		// Write the Action Execution
		ae := &types.ActionExecution{
			ActionId:      actionId,
			CueListId:     cueListId,
			CueId:         cueId,
			ActionStarted: time.Now(),
		}
		err = tx.Insert(TableActionExecution, ae)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	m.registry.Emit(ResourceActionExecution, OperationStarted, MetadataCueListId, cueListId, MetadataCueId, cueId, MetadataActionId, actionId)

	return nil
}

func (m *CueingModel) GetActionExecution(actionId string) (*types.ActionExecution, error) {
	res, err := db.GetFirstDb[types.ActionExecution](m.runtime, TableActionExecution, IndexId, actionId)
	if errors.Is(err, db.ErrItemNotFound) {
		return nil, ErrCueNotFound
	}
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (m *CueingModel) EnumerateActionExecutions(cueId string) ([]types.ActionExecution, error) {
	res, err := db.GetAllDb[types.ActionExecution](m.runtime, TableActionExecution, IndexCueId, cueId)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (m *CueingModel) StopActionExecution(actionId string) error {
	var cueId string
	var cueListId string
	err := db.WithWrite(m.runtime, func(tx *memdb.Txn) error {
		//Check that the Action Execution Exists
		actionExecution, err := db.GetFirstDb[types.ActionExecution](m.persistent, TableActionExecution, IndexId, actionId)
		if errors.Is(err, db.ErrItemNotFound) {
			return nil // idempotent - already finished if it is deleted
		}
		if err != nil {
			return err
		}

		cueId = actionExecution.CueId
		cueListId = actionExecution.CueListId

		return db.DeleteItemFromTxn[types.ActionExecution](tx, TableActionExecution, IndexId, actionId)
	})
	if err != nil {
		return err
	}

	m.registry.Emit(ResourceActionExecution, OperationDeleted, MetadataCueListId, cueListId, MetadataCueId, cueId, MetadataActionId, actionId)

	return nil
}

func (m *CueingModel) StartActionDelayExecution(actionId string) error {
	var updatedCueExec *types.ActionExecution
	err := db.WithWrite(m.runtime, func(tx *memdb.Txn) error {
		cueExec, err := db.GetFirstTxn[types.ActionExecution](tx, TableActionExecution, IndexId, actionId)
		if errors.Is(err, db.ErrItemNotFound) {
			return ErrCueNotFound
		}
		if err != nil {
			return err
		}

		updatedCueExec = &types.ActionExecution{
			CueListId:     cueExec.CueListId,
			CueId:         cueExec.CueId,
			ActionStarted: cueExec.ActionStarted,
			DelayActive:   true,
			DelayStarted:  time.Now(),
		}
		return tx.Insert(TableActionExecution, updatedCueExec)
	})

	// emit updated event
	m.registry.Emit(ResourceActionExecution, OperationUpdated, MetadataCueListId, updatedCueExec.CueListId, MetadataCueId, updatedCueExec.CueId, MetadataActionId, updatedCueExec.ActionId)

	return err
}

func (m *CueingModel) StopActionDelayExecution(actionId string) error {
	var updatedCueExec *types.ActionExecution
	err := db.WithWrite(m.runtime, func(tx *memdb.Txn) error {
		cueExec, err := db.GetFirstTxn[types.ActionExecution](tx, TableCueExecution, IndexId, actionId)
		if errors.Is(err, db.ErrItemNotFound) {
			return ErrCueNotFound
		}
		if err != nil {
			return err
		}

		updatedCueExec = &types.ActionExecution{
			CueListId:     cueExec.CueListId,
			CueId:         cueExec.CueId,
			ActionStarted: cueExec.ActionStarted,
			DelayActive:   false,
			DelayStarted:  cueExec.DelayStarted,
		}
		return tx.Insert(TableCueExecution, updatedCueExec)
	})

	// emit updated event
	m.registry.Emit(ResourceActionExecution, OperationUpdated, MetadataCueListId, updatedCueExec.CueListId, MetadataCueId, updatedCueExec.CueId, MetadataActionId, updatedCueExec.ActionId)

	return err
}
