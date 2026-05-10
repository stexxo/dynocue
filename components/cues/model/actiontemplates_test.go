package model

import (
	"testing"

	"github.com/stexxo/dynocue/components/cues/types"
	"github.com/stretchr/testify/assert"
)

func TestRegisterActionTemplate(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m, _ := NewCueingModel()
		err := m.RegisterActionTemplate(&types.ActionTemplate{
			TemplateId:   "test-template",
			TemplateName: "test-template",
			Subject:      "test-subject",
		})
		assert.NoError(t, err)
	})

	t.Run("template id already exists", func(t *testing.T) {
		m, _ := NewCueingModel()
		err := m.RegisterActionTemplate(&types.ActionTemplate{
			TemplateId:   "test-template",
			TemplateName: "test-template",
			Subject:      "test-subject",
		})
		assert.NoError(t, err)
		err = m.RegisterActionTemplate(&types.ActionTemplate{
			TemplateId:   "test-template",
			TemplateName: "test-template",
			Subject:      "test-subject",
		})
		assert.ErrorIs(t, err, ErrActionTemplateExists)
	})
}

func TestGetActionTemplateById(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m, _ := NewCueingModel()
		err := m.RegisterActionTemplate(&types.ActionTemplate{
			TemplateId:   "test-template",
			TemplateName: "test-template",
			Subject:      "test-subject",
		})
		assert.NoError(t, err)
		template, err := m.GetActionTemplateById("test-template")
		assert.NoError(t, err)
		assert.Equal(t, "test-template", template.TemplateId)
	})

	t.Run("template not found", func(t *testing.T) {
		m, _ := NewCueingModel()
		template, err := m.GetActionTemplateById("non-existent-template")
		assert.ErrorIs(t, err, ErrActionTemplateNotFound)
		assert.Nil(t, template)
	})
}

func TestEnumerateActionTemplates(t *testing.T) {
	m, _ := NewCueingModel()
	err := m.RegisterActionTemplate(&types.ActionTemplate{
		TemplateId:   "test-template",
		TemplateName: "test-template",
		Subject:      "test-subject",
	})
	assert.NoError(t, err)
	templates, err := m.EnumerateActionTemplates()
	assert.NoError(t, err)
	assert.Len(t, templates, 1)
	assert.Equal(t, "test-template", templates[0].TemplateId)
}

func TestDeleteActionTemplate(t *testing.T) {
	m, _ := NewCueingModel()
	err := m.RegisterActionTemplate(&types.ActionTemplate{
		TemplateId:   "test-template",
		TemplateName: "test-template",
		Subject:      "test-subject",
	})
	assert.NoError(t, err)

	_, err = m.GetActionTemplateById("test-template")
	assert.NoError(t, err)

	err = m.DeleteActionTemplateById("test-template")
	assert.NoError(t, err)
	template, err := m.GetActionTemplateById("test-template")
	assert.ErrorIs(t, err, ErrActionTemplateNotFound)
	assert.Nil(t, template)
}
