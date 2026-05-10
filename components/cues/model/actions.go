package model

import (
	"github.com/hashicorp/go-memdb"
	"github.com/stexxo/dynocue/components/cues/types"
	"github.com/stexxo/dynocue/db"
)

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

		num, err := getNextNumber[types.Cue](txn, number, TableActions, IndexNumberPrefix, []any{cueId}, IndexNumber, []any{cueId, number}, func(t *types.Cue) uint {
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

	return a.ActionId, a.Number, nil
}
