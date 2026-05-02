package cues

import (
	"github.com/hashicorp/go-memdb"
)

const (
	TableCueLists = "cuelists"
	TableCues     = "cues"
	TableActions  = "actions"
	TableActionTemplates = "actiontemplates"

	IndexCueListId = "cueListId"
	IndexNumber    = "number"
	IndexCueId     = "cueId"
	IndexActionId  = "actionId"
	IndexActionTemplateId = "actionTemplateId"
)

var persistentSchema = &memdb.DBSchema{
	Tables: map[string]*memdb.TableSchema{
		TableCueLists: {
			Name: TableCueLists,
			Indexes: map[string]*memdb.IndexSchema{
				IndexCueListId: {
					Name:    IndexCueListId,
					Unique:  true,
					Indexer: &memdb.StringFieldIndex{Field: "CueListId"},
				},
				IndexNumber: {
					Name:    IndexNumber,
					Unique:  true,
					Indexer: &memdb.UintFieldIndex{Field: "Number"},
				},
			},
		},
		TableCues: {
			Name: TableCues,
			Indexes: map[string]*memdb.IndexSchema{
				IndexCueListId: {
					Name:    IndexCueListId,
					Unique:  false,
					Indexer: &memdb.UUIDFieldIndex{Field: "CueListId"},
				},
				IndexCueId: {
					Name:    IndexCueId,
					Unique:  true,
					Indexer: &memdb.UUIDFieldIndex{Field: "CueId"},
				},
				IndexNumber: {
					Name:   IndexNumber,
					Unique: true,
					Indexer: &memdb.CompoundIndex{
						Indexes: []memdb.Indexer{
							&memdb.StringFieldIndex{Field: "CueListId"},
							&memdb.UintFieldIndex{Field: "Number"},
						},
					},
				},
			},
		},
		TableActions: {
			Name: TableActions,
			Indexes: map[string]*memdb.IndexSchema{
				IndexCueListId: {
					Name:    IndexCueListId,
					Unique:  false,
					Indexer: &memdb.UUIDFieldIndex{Field: "CueListId"},
				},
				IndexCueId: {
					Name:    IndexCueId,
					Unique:  false,
					Indexer: &memdb.UUIDFieldIndex{Field: "CueId"},
				},
				IndexActionId: {
					Name:    IndexActionId,
					Unique:  false,
					Indexer: &memdb.StringFieldIndex{Field: "ActionId"},
				},
			},
		},
	},
}

var volatileSchema = &memdb.DBSchema{
	Tables: map[string]*memdb.TableSchema{
		TableActionTemplates: {
			Name: TableActionTemplates,
			Indexes: map[string]*memdb.IndexSchema{
				IndexActionTemplateId: {
					Name:    IndexActionTemplateId,
					Unique:  true,
					Indexer: &memdb.StringFieldIndex{Field: "Id"},
				},
			},
		},
	},
}

func (p *Cueing) initiateDatabase() error {
	pdb, err := memdb.NewMemDB(persistentSchema)
	if err != nil {
		return err
	}
	p.db = pdb

	rdb, err := memdb.NewMemDB(volatileSchema)
	if err != nil {
		return err
	}
	p.runtimeDb = rdb

	return nil
}
