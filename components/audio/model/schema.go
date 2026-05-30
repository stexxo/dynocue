package model

import "github.com/hashicorp/go-memdb"

const (
	TableSources = "sources"
	TableFiles   = "files"

	IndexId     = "id"
	IndexFileId = "file_id"
	IndexNumber = "number"
)

var persistentSchema = &memdb.DBSchema{
	Tables: map[string]*memdb.TableSchema{
		TableSources: {
			Name: TableSources,
			Indexes: map[string]*memdb.IndexSchema{
				IndexId: {
					Name:    IndexId,
					Unique:  true,
					Indexer: &memdb.StringFieldIndex{Field: "SourceId"},
				},
				IndexFileId: {
					Name:    IndexFileId,
					Unique:  false,
					Indexer: &memdb.StringFieldIndex{Field: "FileId"},
				},
			},
		},
		TableFiles: {
			Name: TableFiles,
			Indexes: map[string]*memdb.IndexSchema{
				IndexId: {
					Name:    IndexId,
					Unique:  true,
					Indexer: &memdb.StringFieldIndex{Field: "FileId"},
				},
				IndexNumber: {
					Name:    IndexNumber,
					Unique:  true,
					Indexer: &memdb.UintFieldIndex{Field: "Number"},
				},
			},
		},
	},
}

const (
	TableAudioPlayback = "playback"
)

var runtimeSchema = &memdb.DBSchema{
	Tables: map[string]*memdb.TableSchema{
		TableAudioPlayback: {
			Name: TableAudioPlayback,
			Indexes: map[string]*memdb.IndexSchema{
				IndexId: {
					Name:    IndexId,
					Unique:  true,
					Indexer: &memdb.StringFieldIndex{Field: "Id"},
				},
			},
		},
	},
}
