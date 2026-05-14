// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package db

import (
	"errors"

	"github.com/hashicorp/go-memdb"
	"github.com/stexxo/dynocue/util"
)

func WithWrite(db *memdb.MemDB, fn func(*memdb.Txn) error) error {
	txn := db.Txn(true)
	defer txn.Abort()
	if err := fn(txn); err != nil {
		return err
	}
	txn.Commit()
	return nil
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

func UpdateStructInDb[T any](db *memdb.MemDB, table, index string, key any, field string, value interface{}) (*T, error) {
	var out *T
	err := WithWrite(db, func(txn *memdb.Txn) error {
		item, err := GetFirstTxn[T](txn, table, index, key)
		if err != nil {
			return err
		}
		if item == nil {
			return ErrItemNotFound
		}

		// Deep copy of the item to avoid modifying the one in the database
		clone := util.DeepCopyStruct(item)
		err = util.UpdateStructByTag("json", field, value, clone)
		if err != nil {
			return err
		}

		out = clone

		return txn.Insert(table, clone)
	})
	if err != nil {
		return nil, err
	}

	return out, nil
}
