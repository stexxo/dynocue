package model

import (
	"errors"

	"github.com/google/uuid"
	"github.com/hashicorp/go-memdb"
	"github.com/stexxo/dynocue/components/cues/types"
	"github.com/stexxo/dynocue/db"
)

var ErrCueListNotFound = errors.New("cue list not found")

func (m *CueingModel) CreateCueList(number uint, cueListType string) (string, uint, error) {
	m.dbMu.RLock()
	defer m.dbMu.RUnlock()

	cl := types.CueList{
		CueListId:   uuid.NewString(),
		CueListType: cueListType,
	}

	err := db.WithWrite(m.persistent, func(txn *memdb.Txn) error {
		num, err := getNextNumber[types.CueList](txn, number, TableCueLists, IndexNumber, nil, IndexNumber, []any{number}, func(t *types.CueList) uint { return t.Number })
		if err != nil {
			return err
		}

		cl.Number = num
		if err := txn.Insert(TableCueLists, &cl); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return "", 0, err
	}
	m.registry.Emit(ResourceCueList, OperationCreated, MetadataCueListId, cl.CueListId)
	return cl.CueListId, cl.Number, nil
}

func (m *CueingModel) EnumerateCueLists() ([]types.CueList, error) {
	m.dbMu.RLock()
	defer m.dbMu.RUnlock()
	return db.GetAllDb[types.CueList](m.persistent, TableCueLists, IndexNumber)
}

func (m *CueingModel) GetCueListByNumber(number uint) (*types.CueList, error) {
	m.dbMu.RLock()
	defer m.dbMu.RUnlock()
	out, err := db.GetFirstDb[types.CueList](m.persistent, TableCueLists, IndexNumber, number)
	if errors.Is(err, db.ErrItemNotFound) {
		return nil, ErrCueListNotFound
	}
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (m *CueingModel) GetCueListById(id string) (*types.CueList, error) {
	m.dbMu.RLock()
	defer m.dbMu.RUnlock()
	out, err := db.GetFirstDb[types.CueList](m.persistent, TableCueLists, IndexId, id)
	if errors.Is(err, db.ErrItemNotFound) {
		return nil, ErrCueListNotFound
	}
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (m *CueingModel) DeleteCueListById(id string) error {
	m.dbMu.RLock()
	defer m.dbMu.RUnlock()
	err := db.DeleteItemFromDb[types.CueList](m.persistent, TableCueLists, IndexId, id)
	if err != nil {
		return err
	}

	err = m.DeleteAllCuesByCueListId(id)
	if err != nil {
		return err
	}

	m.registry.Emit(ResourceCueList, OperationDeleted, MetadataCueListId, id)
	return nil
}

func (m *CueingModel) UpdateCueListAttribute(id string, field string, value interface{}) error {
	m.dbMu.RLock()
	defer m.dbMu.RUnlock()
	_, err := db.UpdateStructInDb[types.CueList](m.persistent, TableCueLists, IndexId, id, field, value)
	if errors.Is(err, db.ErrItemNotFound) {
		return ErrCueListNotFound
	}
	if err != nil {
		return err
	}
	m.registry.Emit(ResourceCueList, OperationUpdated, MetadataCueListId, id)
	return nil
}
