// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type testStruct struct {
	Name string `msgpack:"name"`
	Age  int    `msgpack:"age"`
}

func TestSetFieldByTag(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		s := testStruct{Name: "Old Name"}
		err := SetFieldByTag(&s, "msgpack", "name", "New Name")
		assert.NoError(t, err)
		assert.Equal(t, "New Name", s.Name)
	})

	t.Run("Not a pointer", func(t *testing.T) {
		s := testStruct{}
		err := SetFieldByTag(s, "msgpack", "name", "New Name")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be a pointer")
	})

	t.Run("Field not found", func(t *testing.T) {
		s := testStruct{}
		err := SetFieldByTag(&s, "msgpack", "nonexistent", "Value")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("Incompatible type", func(t *testing.T) {
		s := testStruct{}
		err := SetFieldByTag(&s, "msgpack", "age", "10")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not a string")
	})
}
