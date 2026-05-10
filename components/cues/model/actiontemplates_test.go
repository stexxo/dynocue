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
