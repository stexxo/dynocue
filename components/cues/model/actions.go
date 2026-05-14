package model

import (
	"errors"
	"fmt"

	"github.com/hashicorp/go-memdb"
	"github.com/stexxo/dynocue/components/cues/types"
	"github.com/stexxo/dynocue/db"
	"github.com/stexxo/dynocue/util"
)

var ErrActionNotFound = errors.New("action not found")

func (m *CueingModel) CreateAction(cueId, templateId string, number uint) (string, uint, error) {
	m.dbMu.RLock()
	defer m.dbMu.RUnlock()

	var a *types.Action

	err := db.WithWrite(m.persistent, func(txn *memdb.Txn) error {
		cue, err := m.GetCueById(cueId)
		if err != nil {
			return err
		}

		template, err := m.GetActionTemplateById(templateId)
		if err != nil {
			return err
		}

		num, err := getNextNumber[types.Action](txn, number, TableActions, IndexNumberPrefix, []any{cueId}, IndexNumber, []any{cueId, number}, func(t *types.Action) uint {
			return t.Number
		})
		if err != nil {
			return err
		}

		a = template.NewAction(cue.CueId, num)

		if err := txn.Insert(TableActions, a); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return "", 0, err
	}

	m.registry.Emit(ResourceAction, OperationCreated, MetadataCueListId, a.CueListId, MetadataCueId, a.CueId, MetadataActionId, a.ActionId)

	return a.ActionId, a.Number, nil
}

func (m *CueingModel) GetActionById(actionId string) (*types.Action, error) {
	m.dbMu.RLock()
	defer m.dbMu.RUnlock()
	out, err := db.GetFirstDb[types.Action](m.persistent, TableActions, IndexId, actionId)
	if errors.Is(err, db.ErrItemNotFound) {
		return nil, ErrActionNotFound
	}
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (m *CueingModel) DeleteAction(actionId string) error {
	m.dbMu.RLock()
	defer m.dbMu.RUnlock()
	a, err := m.GetActionById(actionId)
	if errors.Is(err, ErrActionNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	err = db.DeleteItemFromDb[types.Action](m.persistent, TableActions, IndexId, actionId)
	if err != nil {
		return err
	}
	m.registry.Emit(ResourceAction, OperationDeleted, MetadataCueListId, a.CueListId, MetadataCueId, a.CueId, MetadataActionId, actionId)
	return nil
}

func (m *CueingModel) UpdateAction(actionId string, field string, value any) error {
	m.dbMu.RLock()
	defer m.dbMu.RUnlock()
	a, err := db.UpdateStructInDb[types.Action](m.persistent, TableActions, IndexId, actionId, field, value)
	if errors.Is(err, db.ErrItemNotFound) {
		return ErrActionNotFound
	}
	if err != nil {
		return err
	}
	m.registry.Emit(ResourceAction, OperationUpdated, MetadataCueListId, a.CueListId, MetadataCueId, a.CueId, MetadataActionId, a.ActionId)
	return nil
}

func (m *CueingModel) EnumerateActions(cueId string) ([]types.Action, error) {
	m.dbMu.RLock()
	defer m.dbMu.RUnlock()
	return db.GetAllDb[types.Action](m.persistent, TableActions, IndexNumberPrefix, cueId)
}

func (m *CueingModel) UpdateActionField(actionId string, fieldName string, value any) error {
	m.dbMu.RLock()
	defer m.dbMu.RUnlock()

	var a *types.Action

	err := db.WithWrite(m.persistent, func(txn *memdb.Txn) error {
		original, err := db.GetFirstTxn[types.Action](txn, TableActions, IndexId, actionId)
		if errors.Is(err, db.ErrItemNotFound) {
			return ErrActionNotFound
		}
		if err != nil {
			return err
		}

		// Deep copy of the top level struct and nested slices
		action := util.DeepCopyStruct(original)

		foundField := false
		for i := range action.Fields {
			if action.Fields[i].FieldName == fieldName {
				action.Fields[i].Value = value
				foundField = true
				break
			}
		}

		if !foundField {
			return fmt.Errorf("field %s not found in action %s", fieldName, actionId)
		}

		a = action

		return txn.Insert(TableActions, action)
	})
	if err != nil {
		return err
	}
	m.registry.Emit(ResourceAction, OperationUpdated, MetadataCueListId, a.CueListId, MetadataCueId, a.CueId, MetadataActionId, a.ActionId)
	return nil
}
