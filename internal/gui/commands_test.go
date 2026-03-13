package gui

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommands_OpenCloseShow(t *testing.T) {
	cmds := NewCommands()

	tmpDir, err := os.MkdirTemp("", "commands_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	showPath := filepath.Join(tmpDir, "test_show")

	// Test OpenShow
	path, success := cmds.OpenShow(showPath)
	assert.True(t, success)
	assert.Equal(t, showPath, path)
	assert.NotNil(t, cmds.show)

	// Test Open again (should close previous and open new)
	showPath2 := filepath.Join(tmpDir, "test_show_2")
	path2, success2 := cmds.OpenShow(showPath2)
	assert.True(t, success2)
	assert.Equal(t, showPath2, path2)
	assert.NotNil(t, cmds.show)

	// Test CloseShow
	cmds.CloseShow()
	assert.Nil(t, cmds.show)
}

func TestCommands_LocalShowAliases(t *testing.T) {
	cmds := NewCommands()

	tmpDir, err := os.MkdirTemp("", "commands_alias_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Test OpenLocalShow
	path := filepath.Join(tmpDir, "open_local")
	resPath, success := cmds.OpenLocalShow(path)
	assert.True(t, success)
	assert.Equal(t, path, resPath)
	cmds.CloseShow()

	// Test CreateLocalShow
	path2 := filepath.Join(tmpDir, "create_local")
	resPath2, success2 := cmds.CreateLocalShow(path2)
	assert.True(t, success2)
	assert.Equal(t, path2, resPath2)
	cmds.CloseShow()
}
