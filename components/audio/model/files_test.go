// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package model

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddFileNumbering(t *testing.T) {
	t.Run("auto increments from largest number", func(t *testing.T) {
		m, err := NewAudioModel()
		require.NoError(t, err)

		number, err := m.AddFile("id1", "key1", 0)
		require.NoError(t, err)
		assert.Equal(t, uint(1), number)

		number, err = m.AddFile("id2", "key2", 10)
		require.NoError(t, err)
		assert.Equal(t, uint(10), number)

		number, err = m.AddFile("id3", "key3", 0)
		require.NoError(t, err)
		assert.Equal(t, uint(11), number)

		file, err := m.GetFile("id3")
		require.NoError(t, err)
		assert.Equal(t, uint(11), file.Number)

		file, err = m.GetFileByNumber(10)
		require.NoError(t, err)
		assert.Equal(t, "id2", file.FileId)
	})

	t.Run("specified number must be unique", func(t *testing.T) {
		m, err := NewAudioModel()
		require.NoError(t, err)

		_, err = m.AddFile("id1", "key1", 7)
		require.NoError(t, err)

		_, err = m.AddFile("id2", "key2", 7)
		assert.True(t, errors.Is(err, ErrNumberExists))
	})

	t.Run("enumerates by number", func(t *testing.T) {
		m, err := NewAudioModel()
		require.NoError(t, err)

		_, err = m.AddFile("id5", "key5", 5)
		require.NoError(t, err)
		_, err = m.AddFile("id1", "key1", 1)
		require.NoError(t, err)
		_, err = m.AddFile("id3", "key3", 3)
		require.NoError(t, err)

		files, err := m.EnumerateFiles()
		require.NoError(t, err)

		require.Len(t, files, 3)
		assert.Equal(t, uint(1), files[0].Number)
		assert.Equal(t, uint(3), files[1].Number)
		assert.Equal(t, uint(5), files[2].Number)
	})
}
