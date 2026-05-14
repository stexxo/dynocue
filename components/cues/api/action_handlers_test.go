// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

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

func setupAction(t *testing.T) (*model.CueingModel, *CueingApi, string) {
	m, api := setup(t)
	clId, _, _ := m.CreateCueList(1, types.CueListTypeSequential)
	cueId, _, _ := m.CreateCue(clId, 1)
	_ = m.RegisterActionTemplate(&types.ActionTemplate{
		TemplateId:   "test-template",
		TemplateName: "Test Template",
	})
	return m, api, cueId
}

func TestCreateAction(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		_, api, cueId := setupAction(t)

		req := &CreateActionRequest{
			CueId:      cueId,
			TemplateId: "test-template",
			Number:     1,
		}
		resp, err := api.CreateAction("test-sub", req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.NotEmpty(t, resp.ActionId)
		assert.Equal(t, uint(1), resp.Number)
	})

	t.Run("Error Number Exists", func(t *testing.T) {
		_, api, cueId := setupAction(t)

		req := &CreateActionRequest{
			CueId:      cueId,
			TemplateId: "test-template",
			Number:     1,
		}
		_, _ = api.CreateAction("test-sub", req)

		_, err := api.CreateAction("test-sub", req)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, model.ErrNumberExists))
	})
}

func TestEnumerateActions(t *testing.T) {
	_, api, cueId := setupAction(t)

	_, _ = api.CreateAction("test-sub", &CreateActionRequest{CueId: cueId, TemplateId: "test-template", Number: 1})
	_, _ = api.CreateAction("test-sub", &CreateActionRequest{CueId: cueId, TemplateId: "test-template", Number: 2})

	resp, err := api.EnumerateActions("test-sub", &EnumerateActionsRequest{CueId: cueId})
	assert.NoError(t, err)
	assert.Len(t, resp.Actions, 2)
}

func TestGetActionById(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		_, api, cueId := setupAction(t)
		createResp, _ := api.CreateAction("test-sub", &CreateActionRequest{CueId: cueId, TemplateId: "test-template", Number: 1})

		resp, err := api.GetActionById("test-sub", &GetActionByIdRequest{ActionId: createResp.ActionId})
		assert.NoError(t, err)
		assert.Equal(t, createResp.ActionId, resp.Action.ActionId)
	})

	t.Run("Not Found", func(t *testing.T) {
		_, api, _ := setupAction(t)
		_, err := api.GetActionById("test-sub", &GetActionByIdRequest{ActionId: "non-existent"})
		assert.Error(t, err)
		assert.True(t, errors.Is(err, model.ErrActionNotFound))
	})
}

func TestDeleteAction(t *testing.T) {
	_, api, cueId := setupAction(t)
	createResp, _ := api.CreateAction("test-sub", &CreateActionRequest{CueId: cueId, TemplateId: "test-template", Number: 1})

	_, err := api.DeleteAction("test-sub", &DeleteActionRequest{ActionId: createResp.ActionId})
	assert.NoError(t, err)

	_, err = api.GetActionById("test-sub", &GetActionByIdRequest{ActionId: createResp.ActionId})
	assert.Error(t, err)
}

func TestUpdateAction(t *testing.T) {
	_, api, cueId := setupAction(t)
	createResp, _ := api.CreateAction("test-sub", &CreateActionRequest{CueId: cueId, TemplateId: "test-template", Number: 1})

	_, err := api.UpdateAction("test-sub", &UpdateActionRequest{
		ActionId: createResp.ActionId,
		Field:    "subject",
		Value:    "New Subject",
	})
	assert.NoError(t, err)

	getResp, _ := api.GetActionById("test-sub", &GetActionByIdRequest{ActionId: createResp.ActionId})
	assert.Equal(t, "New Subject", getResp.Action.Subject)
}

func TestRegisterActionApis(t *testing.T) {
	s, nc := testServer()
	defer s.Shutdown()
	defer nc.Close()

	m, _ := model.NewCueingModel()
	js, _ := jetstream.New(nc)
	messenger := messaging.NewMessenger(&messaging.MessengerCfg{
		Conn: nc,
		Js:   js,
	})

	_, err := NewCueingApi(m, nil, messenger, logging.NewNoopLogger())
	require.NoError(t, err)

	clId, _, _ := m.CreateCueList(1, types.CueListTypeSequential)
	cueId, _, _ := m.CreateCue(clId, 1)
	_ = m.RegisterActionTemplate(&types.ActionTemplate{TemplateId: "test", TemplateName: "test"})

	t.Run("Create Registration", func(t *testing.T) {
		req := CreateActionRequest{CueId: cueId, TemplateId: "test", Number: 1}
		env, err := messaging.Request[CreateActionResponse](messenger, CreateActionRequestSubject, req)
		assert.NoError(t, err)
		assert.True(t, env.Success)
	})
}
