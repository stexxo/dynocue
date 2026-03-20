// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package core

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewShow(t *testing.T) {
	t.Run("Create directory if it doesn't exist", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "newshow_test_noexist")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		path := filepath.Join(tmpDir, "myshow")

		s, err := NewShow(path)
		require.NoError(t, err)
		defer s.Close()

		assert.DirExists(t, path)
		assert.FileExists(t, filepath.Join(path, "dynocue.db"))
	})

	t.Run("Fail if path is a file", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "newshow_test_file")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		path := filepath.Join(tmpDir, "not_a_dir")
		err = os.WriteFile(path, []byte("test"), 0644)
		require.NoError(t, err)

		s, err := NewShow(path)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "exists and is not a directory")
		assert.Nil(t, s)
	})

	t.Run("Work if directory exists", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "newshow_test_exists")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		s, err := NewShow(tmpDir)
		require.NoError(t, err)
		defer s.Close()

		assert.FileExists(t, filepath.Join(tmpDir, "dynocue.db"))
	})
}
