package api

import (
	"errors"
	"testing"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/stexxo/dynocue/components/cues/model"
	"github.com/stexxo/dynocue/components/cues/types"
	"github.com/stexxo/dynocue/core/logging"
	"github.com/stexxo/dynocue/core/messaging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupActionTemplate(t *testing.T) (*model.CueingModel, *CueingApi) {
	return setup(t)
}

func TestRegisterActionTemplate(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		_, api := setupActionTemplate(t)

		req := &RegisterActionTemplateRequest{
			Template: types.ActionTemplate{
				TemplateId:   "test-template",
				TemplateName: "Test Template",
			},
		}
		resp, err := api.RegisterActionTemplate("test-sub", req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "test-template", resp.TemplateId)
	})

	t.Run("Error Already Exists", func(t *testing.T) {
		_, api := setupActionTemplate(t)

		req := &RegisterActionTemplateRequest{
			Template: types.ActionTemplate{
				TemplateId:   "test-template",
				TemplateName: "Test Template",
			},
		}
		_, _ = api.RegisterActionTemplate("test-sub", req)

		_, err := api.RegisterActionTemplate("test-sub", req)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, model.ErrActionTemplateExists))
	})
}

func TestEnumerateActionTemplates(t *testing.T) {
	_, api := setupActionTemplate(t)

	_, _ = api.RegisterActionTemplate("test-sub", &RegisterActionTemplateRequest{
		Template: types.ActionTemplate{TemplateId: "t1", TemplateName: "Template 1"},
	})
	_, _ = api.RegisterActionTemplate("test-sub", &RegisterActionTemplateRequest{
		Template: types.ActionTemplate{TemplateId: "t2", TemplateName: "Template 2"},
	})

	resp, err := api.EnumerateActionTemplates("test-sub", &EnumerateActionTemplatesRequest{})
	assert.NoError(t, err)
	assert.Len(t, resp.Templates, 2)
}

func TestGetActionTemplateById(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		_, api := setupActionTemplate(t)
		api.RegisterActionTemplate("test-sub", &RegisterActionTemplateRequest{
			Template: types.ActionTemplate{TemplateId: "t1", TemplateName: "Template 1"},
		})

		resp, err := api.GetActionTemplateById("test-sub", &GetActionTemplateByIdRequest{TemplateId: "t1"})
		assert.NoError(t, err)
		assert.Equal(t, "t1", resp.Template.TemplateId)
	})

	t.Run("Not Found", func(t *testing.T) {
		_, api := setupActionTemplate(t)
		_, err := api.GetActionTemplateById("test-sub", &GetActionTemplateByIdRequest{TemplateId: "non-existent"})
		assert.Error(t, err)
		assert.True(t, errors.Is(err, model.ErrActionTemplateNotFound))
	})
}

func TestDeleteActionTemplate(t *testing.T) {
	_, api := setupActionTemplate(t)
	api.RegisterActionTemplate("test-sub", &RegisterActionTemplateRequest{
		Template: types.ActionTemplate{TemplateId: "t1", TemplateName: "Template 1"},
	})

	_, err := api.DeleteActionTemplate("test-sub", &DeleteActionTemplateRequest{TemplateId: "t1"})
	assert.NoError(t, err)

	_, err = api.GetActionTemplateById("test-sub", &GetActionTemplateByIdRequest{TemplateId: "t1"})
	assert.Error(t, err)
}

func TestRegisterActionTemplateApis(t *testing.T) {
	s, nc := testServer()
	defer s.Shutdown()
	defer nc.Close()

	m, _ := model.NewCueingModel()
	js, _ := jetstream.New(nc)
	messenger := messaging.NewMessenger(&messaging.MessengerCfg{
		Conn: nc,
		Js:   js,
	})

	_, err := NewCueingApi(m, messenger, logging.NewNoopLogger())
	require.NoError(t, err)

	t.Run("Register Registration", func(t *testing.T) {
		req := RegisterActionTemplateRequest{
			Template: types.ActionTemplate{
				TemplateId:   "test",
				TemplateName: "test",
			},
		}
		env, err := messaging.Request[RegisterActionTemplateResponse](messenger, RegisterActionTemplateRequestSubject, req)
		assert.NoError(t, err)
		assert.True(t, env.Success)
		assert.Equal(t, "test", env.Response.TemplateId)
	})
}
