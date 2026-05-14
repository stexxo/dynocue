// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package db

import (
	"errors"

	"github.com/hashicorp/go-memdb"
)

var ErrItemNotFound = errors.New("item not found")

func WithRead[T any](db *memdb.MemDB, fn func(txn *memdb.Txn) (T, error)) (T, error) {
	txn := db.Txn(false)
	defer txn.Abort()
	return fn(txn)
}

func GetLastTxn[T any](txn *memdb.Txn, table string, index string, key ...any) (*T, error) {
	it, err := txn.Last(table, index, key...)
	if err != nil {
		return nil, err
	}
	if it == nil {
		return nil, ErrItemNotFound
	}
	cl := it.(*T)
	return cl, nil
}

func GetAllTxn[T any](txn *memdb.Txn, table string, index string, key ...any) ([]T, error) {
	it, err := txn.Get(table, index, key...)
	if err != nil {
		return nil, err
	}
	var results []T
	for obj := it.Next(); obj != nil; obj = it.Next() {
		o := *obj.(*T)
		results = append(results, o)
	}
	return results, nil
}

func GetAllDb[T any](db *memdb.MemDB, table string, index string, key ...any) ([]T, error) {
	return WithRead[[]T](db, func(txn *memdb.Txn) ([]T, error) {
		return GetAllTxn[T](txn, table, index, key...)
	})
}

func GetFirstTxn[T any](txn *memdb.Txn, table string, index string, key ...any) (*T, error) {
	it, err := txn.First(table, index, key...)
	if err != nil {
		return nil, err
	}
	if it == nil {
		return nil, ErrItemNotFound
	}
	obj := it.(*T)
	return obj, nil
}

func GetFirstDb[T any](db *memdb.MemDB, table string, index string, key ...any) (*T, error) {
	return WithRead[*T](db, func(txn *memdb.Txn) (*T, error) {
		return GetFirstTxn[T](txn, table, index, key...)
	})
}
