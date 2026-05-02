package cues

import (
	"github.com/hashicorp/go-memdb"
)

const (
	TableCueLists = "cuelists"
	TableCues     = "cues"
	TableActions  = "actions"

	IndexCueListId = "cueListId"
	IndexNumber    = "number"
	IndexCueId     = "cueId"
	IndexActionId  = "actionId"
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
					Name:    IndexNumber,
					Unique:  true,
					Indexer: &memdb.UintFieldIndex{Field: "Number"},
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

func (p *Cueing) initiateDatabase() error {
	db, err := memdb.NewMemDB(persistentSchema)
	if err != nil {
		return err
	}
	p.db = db
	return nil
}
