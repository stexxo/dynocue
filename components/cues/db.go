package cues

import (
	"github.com/hashicorp/go-memdb"
)

const (
	TableCueLists        = "cuelists"
	TableCues            = "cues"
	TableActions         = "actions"
	TableActionTemplates = "actiontemplates"

	IndexId     = "id"
	IndexCueId  = "cue_id"
	IndexNumber = "number"

	IndexCueIdPrefix  = "cue_id_prefix"
	IndexNumberPrefix = "number_prefix"
)

var persistentSchema = &memdb.DBSchema{
	Tables: map[string]*memdb.TableSchema{
		TableCueLists: {
			Name: TableCueLists,
			Indexes: map[string]*memdb.IndexSchema{
				IndexId: {
					Name:    IndexId,
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
				IndexId: {
					Name:    IndexId,
					Unique:  true,
					Indexer: &memdb.StringFieldIndex{Field: "CueId"},
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
				IndexId: {
					Name:    IndexId,
					Unique:  true,
					Indexer: &memdb.StringFieldIndex{Field: "ActionId"},
				},
				IndexCueId: {
					Name:    IndexCueId,
					Unique:  false,
					Indexer: &memdb.StringFieldIndex{Field: "CueId"},
				},
			},
		},
	},
}

var runtimeSchema = &memdb.DBSchema{
	Tables: map[string]*memdb.TableSchema{
		TableActionTemplates: {
			Name: TableActionTemplates,
			Indexes: map[string]*memdb.IndexSchema{
				IndexId: {
					Name:    IndexId,
					Unique:  true,
					Indexer: &memdb.StringFieldIndex{Field: "TemplateId"},
				},
			},
		},
	},
}

func (p *Cueing) initiateDatabase() error {
	if p.db == nil {
		pdb, err := memdb.NewMemDB(persistentSchema)
		if err != nil {
			return err
		}
		p.db = pdb
	}

	if p.runtimeDb == nil {
		rdb, err := memdb.NewMemDB(runtimeSchema)
		if err != nil {
			return err
		}
		p.runtimeDb = rdb
	}

	return nil
}
