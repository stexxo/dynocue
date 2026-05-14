// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package model

import (
	"bytes"
	"io"
	"testing"

	"github.com/stexxo/dynocue/components/cues/types"
	"github.com/stexxo/dynocue/db"
	"github.com/stretchr/testify/assert"
)

func TestSerializeAndRestore(t *testing.T) {
	m, cueId := setupActionTest(t)

	// Add some data to ensure we have something in all persistent tables
	// setupActionTest already created a CueList, a Cue and we will create an Action.
	actionId, _, err := m.CreateAction(cueId, "test-template", 0)
	assert.NoError(t, err)

	// 1. SerializeEachTable
	serializedData := make(map[string][]byte)
	err = m.SerializeEachTable(func(name string, reader io.Reader) error {
		data, err := io.ReadAll(reader)
		if err != nil {
			return err
		}
		serializedData[name] = data
		return nil
	})
	assert.NoError(t, err)

	// Verify we got data for all persistent tables
	assert.Contains(t, serializedData, TableCueLists)
	assert.Contains(t, serializedData, TableCues)
	assert.Contains(t, serializedData, TableActions)

	// 2. Restore into a new model
	m2, err := NewCueingModel()
	assert.NoError(t, err)

	// Register the same template in the new model (it's in runtime db, not serialized)
	err = m2.RegisterActionTemplate(&types.ActionTemplate{
		TemplateId:    "test-template",
		TemplateName:  "test-template",
		Subject:       "test-subject",
		SubsystemName: "test-subsystem",
		Fields: []types.ActionTemplateField{
			{FieldName: "test-field-float", FieldLabel: "Test Field Float", DataType: "float", DefaultValue: 0.0},
			{FieldName: "test-field-string", FieldLabel: "Test Field String", DataType: "string", DefaultValue: ""},
		},
	})
	assert.NoError(t, err)

	for name, data := range serializedData {
		err = m2.RestoreTable(name, bytes.NewReader(data))
		assert.NoError(t, err, "failed to restore table %s", name)
	}

	// 3. Verify data consistency
	// Check CueList
	cueLists, err := m2.EnumerateCueLists()
	assert.NoError(t, err)
	assert.Len(t, cueLists, 1)

	// Check Cue
	cues, err := m2.EnumerateCues(cueLists[0].CueListId)
	assert.NoError(t, err)
	assert.Len(t, cues, 1)
	assert.Equal(t, cueId, cues[0].CueId)

	// Check Action
	action, err := m2.GetActionById(actionId)
	assert.NoError(t, err)
	assert.NotNil(t, action)
	assert.Equal(t, actionId, action.ActionId)
	assert.Equal(t, cueId, action.CueId)
}

func TestRestoreTable_NotFound(t *testing.T) {
	m, err := NewCueingModel()
	assert.NoError(t, err)

	err = m.RestoreTable("non-existent-table", bytes.NewReader([]byte{}))
	assert.Error(t, err)
	assert.Equal(t, "table not found", err.Error())
}

func BenchmarkSerializeEachTable(b *testing.B) {
	m, _ := NewCueingModel()

	// Register template so CreateAction works
	_ = m.RegisterActionTemplate(&types.ActionTemplate{
		TemplateId: "test-template",
		Fields:     []types.ActionTemplateField{{FieldName: "f1", DataType: "string"}},
	})

	clId, _, _ := m.CreateCueList(1, types.CueListTypeSequential)
	for i := range 100 {
		cueId, _, _ := m.CreateCue(clId, uint(i+1))
		for j := range 10 {
			_, _, _ = m.CreateAction(cueId, "test-template", uint(j+1))
		}
	}

	b.ResetTimer()
	for b.Loop() {
		_ = m.SerializeEachTable(func(name string, reader io.Reader) error {
			_, _ = io.ReadAll(reader)
			return nil
		})
	}
}

func BenchmarkRestoreTable(b *testing.B) {
	m, _ := NewCueingModel()
	clId, _, _ := m.CreateCueList(1, types.CueListTypeSequential)
	for i := range 100 {
		_, _, _ = m.CreateCue(clId, uint(i+1))
	}

	buf, _ := db.SerializeTable(m.persistent, TableCues)
	data := buf.Bytes()

	b.ResetTimer()
	for b.Loop() {
		m2, _ := NewCueingModel()
		_ = m2.RestoreTable(TableCues, bytes.NewReader(data))
	}
}
