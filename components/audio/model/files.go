package model

import (
	"errors"
	"time"

	"github.com/hashicorp/go-memdb"
	"github.com/stexxo/dynocue/components/audio/types"
	"github.com/stexxo/dynocue/db"
)

var ErrFileNotFound = errors.New("file not found")
var ErrNumberExists = errors.New("number already exists")

func (m *AudioModel) AddFile(id string, key string, number uint) (uint, error) {
	m.dbMu.RLock()
	defer m.dbMu.RUnlock()

	file := types.AudioFile{
		FileId: id,
		Key:    key,
	}

	err := db.WithWrite(m.persistent, func(txn *memdb.Txn) error {
		num, err := getNextFileNumber(txn, number)
		if err != nil {
			return err
		}

		file.Number = num
		return txn.Insert(TableFiles, &file)
	})

	if err != nil {
		return 0, err
	}

	m.registry.Emit(ResourceFile, OperationCreated, MetadataFileId, file.FileId)

	return file.Number, nil
}

func (m *AudioModel) SetFileProperties(fileId string, duration time.Duration, size uint, format string) error {
	m.dbMu.RLock()
	defer m.dbMu.RUnlock()

	err := db.WithWrite(m.persistent, func(txn *memdb.Txn) error {
		f, err := db.GetFirstTxn[types.AudioFile](txn, TableFiles, IndexId, fileId)
		if errors.Is(err, db.ErrItemNotFound) {
			return ErrFileNotFound
		}
		if err != nil {
			return err
		}

		newFile := types.AudioFile{
			FileId:    fileId,
			Number:    f.Number,
			Duration:  duration,
			SizeBytes: size,
			Format:    format,
		}
		return txn.Insert(TableFiles, &newFile)
	})
	if err != nil {
		return err
	}
	m.registry.Emit(ResourceFile, OperationUpdated, MetadataFileId, fileId)
	return nil
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

func (m *AudioModel) GetFileByNumber(number uint) (*types.AudioFile, error) {
	m.dbMu.RLock()
	defer m.dbMu.RUnlock()
	item, err := db.GetFirstDb[types.AudioFile](m.persistent, TableFiles, IndexNumber, number)
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
	return db.GetAllDb[types.AudioFile](m.persistent, TableFiles, IndexNumber)
}

func getNextFileNumber(txn *memdb.Txn, number uint) (uint, error) {
	if number == 0 {
		last, err := db.GetLastTxn[types.AudioFile](txn, TableFiles, IndexNumber)
		if errors.Is(err, db.ErrItemNotFound) {
			return 1, nil
		}
		if err != nil {
			return 0, err
		}
		return last.Number + 1, nil
	}

	existing, err := txn.First(TableFiles, IndexNumber, number)
	if err != nil {
		return 0, err
	}
	if existing != nil {
		return 0, ErrNumberExists
	}
	return number, nil
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
