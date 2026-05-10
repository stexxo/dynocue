package model

import (
	"testing"

	"github.com/stexxo/dynocue/components/cues/types"
	"github.com/stretchr/testify/assert"
)

func TestCreateAction(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		m, _ := NewCueingModel()

		//Create CueList
		clId, _, err := m.CreateCueList(1, types.CueListTypeSequential)
		assert.NoError(t, err)

		//Create Cue
		cueId, _, err := m.CreateCue(clId, 0)
		assert.NoError(t, err)

		//Create Action Template
		err = m.RegisterActionTemplate(&types.ActionTemplate{
			TemplateId:    "test-template",
			TemplateName:  "test-template",
			Subject:       "test-subject",
			SubsystemName: "test-subsystem",
			Fields: []types.ActionTemplateField{
				{FieldName: "test-field-float", FieldLabel: "Test Field Float", DataType: "float", DefaultValue: 0.0},
				{FieldName: "test-field-string", FieldLabel: "Test Field String", DataType: "string", DefaultValue: ""},
			},
		})

		// Create Action
		id, number, err := m.CreateAction(cueId, "test-template", 0)
		assert.NoError(t, err)
		assert.NotEmpty(t, id)
		assert.Equal(t, uint(1), number)

		t.Run("number already exists", func(t *testing.T) {
			id, number, err := m.CreateAction(cueId, "test-template", 1)
			assert.ErrorIs(t, err, ErrNumberExists)
			assert.Empty(t, id)
			assert.Empty(t, number)
		})
	})

	t.Run("Cue not Found", func(t *testing.T) {
		m, _ := NewCueingModel()
		_, _, err := m.CreateAction("notreal", "notreal", 0)
		assert.ErrorIs(t, err, ErrCueNotFound)
	})

	t.Run("Template Not Found", func(t *testing.T) {
		m, _ := NewCueingModel()

		//Create CueList
		clId, _, err := m.CreateCueList(1, types.CueListTypeSequential)
		assert.NoError(t, err)

		//Create Cue
		cueId, _, err := m.CreateCue(clId, 0)
		assert.NoError(t, err)

		// Create Action
		id, number, err := m.CreateAction(cueId, "test-template", 0)
		assert.ErrorIs(t, err, ErrActionTemplateNotFound)
		assert.Empty(t, id)
		assert.Empty(t, number)
	})
}
