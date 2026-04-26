// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package util

import (
	"cmp"
	"encoding/json"
	"errors"
	"math"
	"slices"
	"sync"
)

type Number interface {
	Num() float64
	SetNum(float64)
}

type NumberedSlice[T Number] struct {
	mu   sync.RWMutex
	data []T
}

func NewNumberedSlice[T Number]() *NumberedSlice[T] {
	return &NumberedSlice[T]{
		data: make([]T, 0),
	}
}

func (o *NumberedSlice[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Data []T `json:"data"`
	}{Data: o.data})
}

func (o *NumberedSlice[T]) UnmarshalJSON(data []byte) error {
	decoded := &struct {
		Data []T `json:"data"`
	}{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	o.data = decoded.Data
	return nil
}

func (o *NumberedSlice[T]) Len() int {
	return len(o.data)
}

func (o *NumberedSlice[T]) Add(in T) bool {
	o.mu.Lock()
	defer o.mu.Unlock()

	var index int

	if in.Num() == 0 {
		if o.Len() == 0 {
			in.SetNum(1)
			index = 0
		} else {
			c := o.data[o.Len()-1]
			in.SetNum(math.Floor(c.Num()) + 1)
			index = o.Len()
		}
	} else {
		i, found := slices.BinarySearchFunc(o.data, in.Num(), func(a T, b float64) int {
			return cmp.Compare(a.Num(), b)
		})
		if found {
			return false
		}
		index = i
	}

	o.data = slices.Insert(o.data, index, in)
	return true
}

var ErrNotFound = errors.New("item not found")
var ErrExists = errors.New("item already exists")

func (o *NumberedSlice[T]) MoveFunc(fn func(T) bool, newNumber float64) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	originalIdx := slices.IndexFunc(o.data, fn)
	if originalIdx == -1 {
		return ErrNotFound
	}

	_, found := slices.BinarySearchFunc(o.data, newNumber, func(a T, b float64) int {
		return cmp.Compare(a.Num(), b)
	})
	if found {
		return ErrExists
	}

	o.data[originalIdx].SetNum(newNumber)

	slices.SortFunc(o.data, func(a, b T) int {
		return cmp.Compare(a.Num(), b.Num())
	})

	return nil
}

func (o *NumberedSlice[T]) RemoveFunc(fn func(T) bool) {
	o.mu.Lock()
	defer o.mu.Unlock()

	i := slices.IndexFunc(o.data, fn)
	if i == -1 {
		return
	}
	o.data = slices.Delete(o.data, i, i+1)
}

func (o *NumberedSlice[T]) GetByNumber(num float64) *T {
	o.mu.RLock()
	defer o.mu.RUnlock()

	idx, found := slices.BinarySearchFunc(o.data, num, func(a T, b float64) int {
		return cmp.Compare(a.Num(), b)
	})

	if !found {
		return nil
	}

	return &o.data[idx]
}

func (o *NumberedSlice[T]) GetNextByNumber(current float64) *T {
	o.mu.RLock()
	defer o.mu.RUnlock()

	idx, found := slices.BinarySearchFunc(o.data, current, func(a T, b float64) int {
		return cmp.Compare(a.Num(), b)
	})
	if !found {
		return nil
	}

	if idx == len(o.data)-1 {
		return nil
	}

	return &o.data[idx+1]
}

func (o *NumberedSlice[T]) GetFunc(fn func(T) bool) *T {
	o.mu.RLock()
	defer o.mu.RUnlock()

	i := slices.IndexFunc(o.data, fn)
	if i == -1 {
		return nil
	}

	return &o.data[i]
}

func (o *NumberedSlice[T]) ForEach(fn func(T)) {
	o.mu.RLock()
	defer o.mu.RUnlock()
	for _, item := range o.data {
		fn(item)
	}
}

func (o *NumberedSlice[T]) WithValues(fn func([]T)) {
	o.mu.RLock()
	defer o.mu.RUnlock()
	fn(o.data)
}
