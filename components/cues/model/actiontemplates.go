package model

import (
	"errors"

	"github.com/hashicorp/go-memdb"
	"github.com/stexxo/dynocue/components/cues/types"
	"github.com/stexxo/dynocue/db"
)

var ErrActionTemplateExists = errors.New("action template already exists")
var ErrActionTemplateNotFound = errors.New("action template not found")

func (m *CueingModel) RegisterActionTemplate(template *types.ActionTemplate) error {
	m.dbMu.RLock()
	defer m.dbMu.RUnlock()
	err := db.WithWrite(m.runtime, func(txn *memdb.Txn) error {
		temp, err := txn.First(TableActionTemplates, IndexId, template.TemplateId)
		if err != nil {
			return err
		}
		if temp != nil {
			return ErrActionTemplateExists
		}
		return txn.Insert(TableActionTemplates, template)
	})
	if err != nil {
		return err
	}
	m.registry.Emit(ResourceActionTemplate, OperationCreated, template.TemplateId)
	return nil
}

func (m *CueingModel) GetActionTemplateById(templateId string) (*types.ActionTemplate, error) {
	m.dbMu.RLock()
	defer m.dbMu.RUnlock()
	item, err := db.GetFirstDb[types.ActionTemplate](m.runtime, TableActionTemplates, IndexId, templateId)
	if errors.Is(err, db.ErrItemNotFound) {
		return nil, errors.Join(err, ErrActionTemplateNotFound)
	}
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (m *CueingModel) EnumerateActionTemplates() ([]types.ActionTemplate, error) {
	m.dbMu.RLock()
	defer m.dbMu.RUnlock()
	return db.GetAllDb[types.ActionTemplate](m.runtime, TableActionTemplates, IndexId)
}

func (m *CueingModel) DeleteActionTemplateById(templateId string) error {
	m.dbMu.RLock()
	defer m.dbMu.RUnlock()
	err := db.DeleteItemFromDb[types.ActionTemplate](m.runtime, TableActionTemplates, IndexId, templateId)
	if err != nil {
		return err
	}
	m.registry.Emit(ResourceActionTemplate, OperationDeleted, templateId)
	return nil
}
