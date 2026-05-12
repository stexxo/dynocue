package model

import (
	"testing"

	"github.com/stexxo/dynocue/components/cues/types"
	"github.com/stretchr/testify/assert"
)

func setupActionTest(t *testing.T) (m *CueingModel, cueId string) {
	m, _ = NewCueingModel()

	// Create CueList
	clId, _, err := m.CreateCueList(1, types.CueListTypeSequential)
	assert.NoError(t, err)

	// Create Cue
	cueId, _, err = m.CreateCue(clId, 0)
	assert.NoError(t, err)

	// Create Action Template
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
	assert.NoError(t, err)

	return m, cueId
}

func TestCreateAction(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		m, cueId := setupActionTest(t)

		// Create Action
		id, number, err := m.CreateAction(cueId, "test-template", 0)
		assert.NoError(t, err)
		assert.NotEmpty(t, id)
		assert.Equal(t, uint(1), number)
	})

	t.Run("number already exists", func(t *testing.T) {
		m, cueId := setupActionTest(t)

		// Create first action
		_, _, err := m.CreateAction(cueId, "test-template", 1)
		assert.NoError(t, err)

		// Try to create second action with same number
		id2, number2, err := m.CreateAction(cueId, "test-template", 1)
		assert.ErrorIs(t, err, ErrNumberExists)
		assert.Empty(t, id2)
		assert.Empty(t, number2)
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

func TestGetActionById(t *testing.T) {
	t.Parallel()
	m, cueId := setupActionTest(t)

	id, _, err := m.CreateAction(cueId, "test-template", 0)
	assert.NoError(t, err)

	t.Run("Success", func(t *testing.T) {
		action, err := m.GetActionById(id)
		assert.NoError(t, err)
		assert.Equal(t, id, action.ActionId)
		assert.Equal(t, cueId, action.CueId)
	})

	t.Run("Not Found", func(t *testing.T) {
		action, err := m.GetActionById("non-existent")
		assert.ErrorIs(t, err, ErrActionNotFound)
		assert.Nil(t, action)
	})
}

func TestUpdateAction(t *testing.T) {
	t.Parallel()
	m, cueId := setupActionTest(t)

	id, _, err := m.CreateAction(cueId, "test-template", 0)
	assert.NoError(t, err)

	t.Run("Success", func(t *testing.T) {
		err := m.UpdateAction(id, "label", "new label")
		assert.NoError(t, err)

		action, err := m.GetActionById(id)
		assert.NoError(t, err)
		assert.Equal(t, "new label", action.Label)
	})

	t.Run("Not Found", func(t *testing.T) {
		err := m.UpdateAction("non-existent", "label", "new label")
		assert.ErrorIs(t, err, ErrActionNotFound)
	})
}

func TestUpdateActionField(t *testing.T) {
	t.Parallel()
	m, cueId := setupActionTest(t)

	id, _, err := m.CreateAction(cueId, "test-template", 0)
	assert.NoError(t, err)

	t.Run("Success", func(t *testing.T) {
		err := m.UpdateActionField(id, "test-field-float", 1.5)
		assert.NoError(t, err)

		action, err := m.GetActionById(id)
		assert.NoError(t, err)
		found := false
		for _, f := range action.Fields {
			if f.FieldName == "test-field-float" {
				assert.Equal(t, 1.5, f.Value)
				found = true
			}
		}
		assert.True(t, found)
	})

	t.Run("Action Not Found", func(t *testing.T) {
		err := m.UpdateActionField("non-existent", "test-field-float", 1.5)
		assert.ErrorIs(t, err, ErrActionNotFound)
	})

	t.Run("Field Not Found", func(t *testing.T) {
		err := m.UpdateActionField(id, "non-existent-field", 1.5)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "field non-existent-field not found")
	})
}

func TestDeleteAction(t *testing.T) {
	t.Parallel()
	m, cueId := setupActionTest(t)

	id, _, err := m.CreateAction(cueId, "test-template", 0)
	assert.NoError(t, err)

	t.Run("Success", func(t *testing.T) {
		err := m.DeleteAction(id)
		assert.NoError(t, err)

		action, err := m.GetActionById(id)
		assert.ErrorIs(t, err, ErrActionNotFound)
		assert.Nil(t, action)
	})

	t.Run("Not Found", func(t *testing.T) {
		err := m.DeleteAction("non-existent")
		assert.NoError(t, err)
	})
}
