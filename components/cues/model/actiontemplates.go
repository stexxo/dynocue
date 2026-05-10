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
	return nil
}

func (m *CueingModel) GetActionTemplateById(templateId string) (*types.ActionTemplate, error) {
	item, err := db.GetFirstDb[types.ActionTemplate](m.runtime, TableActionTemplates, IndexId, templateId)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, ErrActionTemplateNotFound
	}
	return item, nil
}
