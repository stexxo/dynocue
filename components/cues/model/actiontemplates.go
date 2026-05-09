package model

import (
	"errors"

	"github.com/hashicorp/go-memdb"
	"github.com/stexxo/dynocue/components/cues/types"
	"github.com/stexxo/dynocue/db"
)

var ErrActionTemplateExists = errors.New("action template already exists")

func (m *CueingModel) RegisterActionTemplate(template types.ActionTemplate) error {
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
