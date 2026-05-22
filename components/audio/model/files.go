package model

import (
	"errors"

	"github.com/google/uuid"
	"github.com/hashicorp/go-memdb"
	"github.com/stexxo/dynocue/components/audio/types"
	"github.com/stexxo/dynocue/db"
)

var ErrFileNotFound = errors.New("file not found")

func (m *AudioModel) AddFile(key string) (string, error) {
	m.dbMu.RLock()
	defer m.dbMu.RUnlock()

	file := types.AudioFile{
		FileId: uuid.NewString(),
		Key:    key,
	}

	err := db.WithWrite(m.persistent, func(txn *memdb.Txn) error {
		return txn.Insert(TableFiles, file)
	})

	if err != nil {
		return "", err
	}

	m.registry.Emit(ResourceFile, OperationCreated, MetadataFileId, file.FileId)

	return file.FileId, nil
}

func (m *AudioModel) GetFile(fileId string) (*types.AudioFile, error) {
	m.dbMu.RLock()
	defer m.dbMu.RUnlock()
	item, err := db.GetFirstDb[types.AudioFile](m.persistent, TableFiles, IndexId, fileId)
	if errors.Is(err, db.ErrItemNotFound) {
		return nil, ErrFileNotFound
	}
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (m *AudioModel) EnumerateFiles() ([]types.AudioFile, error) {
	m.dbMu.RLock()
	defer m.dbMu.RUnlock()
	return db.GetAllDb[types.AudioFile](m.persistent, TableFiles, IndexId)
}

func (m *AudioModel) UpdateFile(fileId string, field string, value any) error {
	m.dbMu.RLock()
	defer m.dbMu.RUnlock()
	_, err := db.UpdateStructDb[types.AudioFile](m.persistent, TableFiles, IndexId, fileId, field, value)
	if errors.Is(err, db.ErrItemNotFound) {
		return ErrFileNotFound
	}
	if err != nil {
		return err
	}
	m.registry.Emit(ResourceFile, OperationUpdated, MetadataFileId, fileId)
	return nil
}

func (m *AudioModel) DeleteFile(fileId string) error {
	m.dbMu.RLock()
	defer m.dbMu.RUnlock()
	sourcesUpdated := []string{}
	err := db.WithWrite(m.persistent, func(txn *memdb.Txn) error {
		err := db.DeleteItemFromTxn[types.AudioFile](txn, TableFiles, IndexId, fileId)
		if err != nil {
			return err
		}

		sources, err := db.GetAllTxn[types.AudioSource](txn, TableSources, IndexFileId, fileId)
		if err != nil {
			return err
		}
		for _, source := range sources {
			_, err := db.UpdateStructTxn[types.AudioSource](txn, TableSources, IndexId, source.SourceId, "FileId", "")
			if err != nil {
				return err
			}
			sourcesUpdated = append(sourcesUpdated, source.SourceId)
		}
		return nil
	})
	if err != nil {
		return err
	}

	m.registry.Emit(ResourceFile, OperationDeleted, MetadataFileId, fileId)
	for _, sourceId := range sourcesUpdated {
		m.registry.Emit(ResourceSource, OperationUpdated, MetadataSourceId, sourceId)
	}

	return nil
}
