// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package model

import (
	"github.com/hashicorp/go-memdb"
)

const (
	TableCueLists        = "cuelists"
	TableCues            = "cues"
	TableActions         = "actions"
	TableActionTemplates = "actiontemplates"
	TableCueListPlayback = "cueplayback"

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
				IndexNumber: {
					Name:   IndexNumber,
					Unique: true,
					Indexer: &memdb.CompoundIndex{
						Indexes: []memdb.Indexer{
							&memdb.StringFieldIndex{Field: "CueId"},
							&memdb.UintFieldIndex{Field: "Number"},
						},
					},
				},
				IndexNumberPrefix: {
					Name: IndexNumberPrefix,
					Indexer: &memdb.CompoundIndex{
						Indexes: []memdb.Indexer{
							&memdb.StringFieldIndex{Field: "CueId"},
						},
					},
				},
				IndexCueIdPrefix: {
					Name:    IndexCueIdPrefix,
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
		TableCueListPlayback: {
			Name: TableCueListPlayback,
			Indexes: map[string]*memdb.IndexSchema{
				IndexId: {
					Name:    IndexId,
					Unique:  true,
					Indexer: &memdb.StringFieldIndex{Field: "CueListId"},
				},
			},
		},
	},
}
