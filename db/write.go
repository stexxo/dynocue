package db

import (
	"errors"

	"github.com/hashicorp/go-memdb"
	"github.com/stexxo/dynocue/util"
)

func WithWrite(db *memdb.MemDB, fn func(*memdb.Txn) error) error {
	txn := db.Txn(true)
	defer txn.Abort()
	return fn(txn)
}

func DeleteItemFromTxn[T any](tx *memdb.Txn, table string, index string, key any) error {
	item, err := GetFirstTxn[T](tx, table, index, key)
	if errors.Is(err, ErrItemNotFound) {
		return nil
	}
	if err != nil {
		return err
	}

	err = tx.Delete(table, item)
	if err != nil {
		return err
	}
	return nil
}

func DeleteItemFromDb[T any](db *memdb.MemDB, table string, index string, key any) error {
	return WithWrite(db, func(txn *memdb.Txn) error {
		return DeleteItemFromTxn[T](txn, table, index, key)
	})
}

func UpdateStructInDb(db *memdb.MemDB, table, index string, key any, field string, value interface{}) error {
	err := WithWrite(db, func(txn *memdb.Txn) error {
		item, err := GetFirstTxn[any](txn, table, index, key)
		if err != nil {
			return err
		}

		// Deep copy of the item to avoid modifying the one in the database
		clone := util.DeepCopyStruct(item)
		err = util.UpdateStructByTag("json", field, value, clone)
		if err != nil {
			return err
		}

		return txn.Insert(table, clone)
	})
	return err
}
