package model

import (
	"errors"

	"github.com/google/uuid"
	"github.com/hashicorp/go-memdb"
	"github.com/stexxo/dynocue/components/audio/types"
	"github.com/stexxo/dynocue/db"
)

var ErrSourceNotFound = errors.New("source not found")

func (m *AudioModel) CreateSource(fileId string) (string, error) {

	source := &types.AudioSource{
		SourceId: uuid.NewString(),
		FileId:   fileId,
	}

	err := db.WithWrite(m.persistent, func(txn *memdb.Txn) error {
		if fileId != "" { // If a fileId is provided, check that it exists
			_, err := db.GetFirstTxn[types.AudioFile](txn, TableFiles, IndexId, fileId)
			if errors.Is(err, db.ErrItemNotFound) {
				return ErrFileNotFound
			}
			if err != nil {
				return err
			}
		}
		return txn.Insert(TableSources, source)
	})

	if err != nil {
		return "", err
	}

	m.registry.Emit(ResourceSource, OperationCreated, MetadataSourceId, source.SourceId)

	return source.SourceId, nil
}

func (m *AudioModel) GetSource(sourceId string) (*types.AudioSource, error) {
	m.dbMu.RLock()
	defer m.dbMu.RUnlock()
	source, err := db.GetFirstDb[types.AudioSource](m.persistent, TableSources, IndexId, sourceId)
	if errors.Is(err, db.ErrItemNotFound) {
		return nil, ErrSourceNotFound
	}
	if err != nil {
		return nil, err
	}
	return source, nil
}

func (m *AudioModel) EnumerateSources() ([]types.AudioSource, error) {
	m.dbMu.RLock()
	defer m.dbMu.RUnlock()
	return db.GetAllDb[types.AudioSource](m.persistent, TableSources, IndexId)
}

func (m *AudioModel) EnumerateSourcesByFileId(fileId string) ([]types.AudioSource, error) {
	m.dbMu.RLock()
	defer m.dbMu.RUnlock()
	return db.GetAllDb[types.AudioSource](m.persistent, TableSources, IndexFileId, fileId)
}

func (m *AudioModel) DeleteSource(sourceId string) error {
	m.dbMu.RLock()
	defer m.dbMu.RUnlock()
	err := db.DeleteItemFromDb[types.AudioSource](m.persistent, TableSources, IndexId, sourceId)
	if err != nil {
		return err
	}
	m.registry.Emit(ResourceSource, OperationDeleted, MetadataSourceId, sourceId)
	return nil
}

func (m *AudioModel) UpdateSourceAttribute(sourceId string, field string, value interface{}) error {
	m.dbMu.RLock()
	defer m.dbMu.RUnlock()
	_, err := db.UpdateStructDb[types.AudioSource](m.persistent, TableSources, IndexId, sourceId, field, value)
	if errors.Is(err, db.ErrItemNotFound) {
		return ErrSourceNotFound
	}
	if err != nil {
		return err
	}
	m.registry.Emit(ResourceSource, OperationUpdated, MetadataSourceId, sourceId)
	return nil
}
