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

func TestCreateCue(t *testing.T) {
	t.Run("Success with specific number", func(t *testing.T) {
		m, api := setup(t)
		clId, _, _ := m.CreateCueList(1, types.CueListTypeSequential)

		req := &CreateCueRequest{
			CueListId: clId,
			Number:    10,
		}
		resp, err := api.CreateCue("test-sub", req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.NotEmpty(t, resp.CueId)
		assert.Equal(t, uint(10), resp.Number)

		// Verify it's in the model
		c, err := m.GetCueById(resp.CueId)
		assert.NoError(t, err)
		assert.Equal(t, uint(10), c.Number)
		assert.Equal(t, clId, c.CueListId)
	})

	t.Run("Success with auto-increment (number 0)", func(t *testing.T) {
		m, api := setup(t)
		clId, _, _ := m.CreateCueList(1, types.CueListTypeSequential)

		req := &CreateCueRequest{
			CueListId: clId,
			Number:    0,
		}

		resp, err := api.CreateCue("test-sub", req)
		assert.NoError(t, err)
		assert.Equal(t, uint(1), resp.Number)

		resp2, err := api.CreateCue("test-sub", req)
		assert.NoError(t, err)
		assert.Equal(t, uint(2), resp2.Number)
	})

	t.Run("Error number exists", func(t *testing.T) {
		m, api := setup(t)
		clId, _, _ := m.CreateCueList(1, types.CueListTypeSequential)

		req := &CreateCueRequest{
			CueListId: clId,
			Number:    5,
		}
		_, err := api.CreateCue("test-sub", req)
		assert.NoError(t, err)

		// Try to create again with same number
		resp, err := api.CreateCue("test-sub", req)
		assert.Error(t, err)
		assert.Nil(t, resp)

		assert.True(t, errors.Is(err, model.ErrNumberExists))

		var friendlyErr *messaging.FriendlyError
		assert.True(t, errors.As(err, &friendlyErr))
		assert.Equal(t, CueNumberExists, friendlyErr.FriendlyErr)
	})
}

func TestEnumerateCues(t *testing.T) {
	t.Run("Success with multiple cues", func(t *testing.T) {
		m, api := setup(t)
		clId, _, _ := m.CreateCueList(1, types.CueListTypeSequential)

		// Seed some data
		_, _, _ = m.CreateCue(clId, 1)
		_, _, _ = m.CreateCue(clId, 2)
		_, _, _ = m.CreateCue(clId, 5)

		resp, err := api.EnumerateCues("test-sub", &EnumerateCuesRequest{CueListId: clId})
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp.Cues, 3)

		// Verify numbers (should be sorted by number if model supports it)
		assert.Equal(t, uint(1), resp.Cues[0].Number)
		assert.Equal(t, uint(2), resp.Cues[1].Number)
		assert.Equal(t, uint(5), resp.Cues[2].Number)
	})

	t.Run("Success with empty list", func(t *testing.T) {
		m, api := setup(t)
		clId, _, _ := m.CreateCueList(1, types.CueListTypeSequential)

		resp, err := api.EnumerateCues("test-sub", &EnumerateCuesRequest{CueListId: clId})
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Empty(t, resp.Cues)
	})
}

func TestGetCueByNumber(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m, api := setup(t)
		clId, _, _ := m.CreateCueList(1, types.CueListTypeSequential)
		_, num, _ := m.CreateCue(clId, 10)

		resp, err := api.GetCueByNumber("test-sub", &GetCueByNumberRequest{CueListId: clId, Number: float64(num)})
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, uint(10), resp.Cue.Number)
		assert.Equal(t, clId, resp.Cue.CueListId)
	})

	t.Run("Not Found", func(t *testing.T) {
		m, api := setup(t)
		clId, _, _ := m.CreateCueList(1, types.CueListTypeSequential)

		resp, err := api.GetCueByNumber("test-sub", &GetCueByNumberRequest{CueListId: clId, Number: 99})
		assert.Error(t, err)
		assert.Nil(t, resp)

		assert.True(t, errors.Is(err, model.ErrCueNotFound))
		if ferr, ok := errors.AsType[*messaging.FriendlyError](err); ok {
			assert.Equal(t, CueNotFound, ferr.FriendlyErr)
		} else {
			t.Errorf("Expected FriendlyError")
		}
	})
}

func TestGetCueById(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m, api := setup(t)
		clId, _, _ := m.CreateCueList(1, types.CueListTypeSequential)
		cueId, _, _ := m.CreateCue(clId, 10)

		resp, err := api.GetCueById("test-sub", &GetCueByIdRequest{CueId: cueId})
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, cueId, resp.Cue.CueId)
		assert.Equal(t, uint(10), resp.Cue.Number)
	})

	t.Run("Not Found", func(t *testing.T) {
		_, api := setup(t)

		resp, err := api.GetCueById("test-sub", &GetCueByIdRequest{CueId: "non-existent"})
		assert.Error(t, err)
		assert.Nil(t, resp)

		assert.True(t, errors.Is(err, model.ErrCueNotFound))
		if ferr, ok := errors.AsType[*messaging.FriendlyError](err); ok {
			assert.Equal(t, CueNotFound, ferr.FriendlyErr)
		} else {
			t.Errorf("Expected FriendlyError")
		}
	})
}

func TestDeleteCue(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m, api := setup(t)
		clId, _, _ := m.CreateCueList(1, types.CueListTypeSequential)
		cueId, _, _ := m.CreateCue(clId, 10)

		resp, err := api.DeleteCue("test-sub", &DeleteCueRequest{CueId: cueId})
		assert.NoError(t, err)
		assert.NotNil(t, resp)

		// Verify it's gone
		_, err = m.GetCueById(cueId)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, model.ErrCueNotFound))
	})
}

func TestUpdateCueAttributes(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m, api := setup(t)
		clId, _, _ := m.CreateCueList(1, types.CueListTypeSequential)
		cueId, _, _ := m.CreateCue(clId, 10)

		req := &UpdateCueAttributesRequest{
			CueId: cueId,
			Field: "label",
			Value: "New Label",
		}
		resp, err := api.UpdateCueAttributes("test-sub", req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)

		// Verify change
		c, err := m.GetCueById(cueId)
		assert.NoError(t, err)
		assert.Equal(t, "New Label", c.Label)
	})

	t.Run("Not Found", func(t *testing.T) {
		_, api := setup(t)

		req := &UpdateCueAttributesRequest{
			CueId: "non-existent",
			Field: "label",
			Value: "New Label",
		}
		resp, err := api.UpdateCueAttributes("test-sub", req)
		assert.Error(t, err)
		assert.Nil(t, resp)

		assert.True(t, errors.Is(err, model.ErrCueNotFound))
		if ferr, ok := errors.AsType[*messaging.FriendlyError](err); ok {
			assert.Equal(t, CueNotFound, ferr.FriendlyErr)
		} else {
			t.Errorf("Expected FriendlyError")
		}
	})
}

func TestRegisterCueApis(t *testing.T) {
	s, nc := testServer()
	defer s.Shutdown()
	defer nc.Close()

	m, _ := model.NewCueingModel()
	js, _ := jetstream.New(nc)
	messenger := messaging.NewMessenger(&messaging.MessengerCfg{
		Conn: nc,
		Js:   js,
	})

	// CreateCueingApi will call registerCueApis
	api, err := NewCueingApi(m, nil, nil, messenger, logging.NewNoopLogger())
	require.NoError(t, err)
	require.NotNil(t, api)

	clId, _, _ := m.CreateCueList(1, types.CueListTypeSequential)

	t.Run("Create Registration", func(t *testing.T) {
		req := CreateCueRequest{CueListId: clId, Number: 10}
		env, err := messaging.Request[CreateCueResponse](messenger, CreateCueRequestSubject, req)
		assert.NoError(t, err)
		assert.True(t, env.Success)
		assert.NotEmpty(t, env.Response.CueId)
	})

	t.Run("Enumerate Registration", func(t *testing.T) {
		req := EnumerateCuesRequest{CueListId: clId}
		env, err := messaging.Request[EnumerateCuesResponse](messenger, EnumerateCuesRequestSubject, req)
		assert.NoError(t, err)
		assert.True(t, env.Success)
	})

	t.Run("GetByNumber Registration", func(t *testing.T) {
		cueId, _, _ := m.CreateCue(clId, 20)
		req := GetCueByNumberRequest{CueListId: clId, Number: 20}
		env, err := messaging.Request[GetCueByNumberResponse](messenger, GetCueByNumberRequestSubject, req)
		assert.NoError(t, err)
		assert.True(t, env.Success)
		assert.Equal(t, cueId, env.Response.Cue.CueId)
	})

	t.Run("GetById Registration", func(t *testing.T) {
		cueId, _, _ := m.CreateCue(clId, 30)
		req := GetCueByIdRequest{CueId: cueId}
		env, err := messaging.Request[GetCueByIdResponse](messenger, GetCueByIdRequestSubject, req)
		assert.NoError(t, err)
		assert.True(t, env.Success)
		assert.Equal(t, uint(30), env.Response.Cue.Number)
	})

	t.Run("Delete Registration", func(t *testing.T) {
		cueId, _, _ := m.CreateCue(clId, 40)
		req := DeleteCueRequest{CueId: cueId}
		env, err := messaging.Request[DeleteCueResponse](messenger, DeleteCueRequestSubject, req)
		assert.NoError(t, err)
		assert.True(t, env.Success)
	})

	t.Run("UpdateAttributes Registration", func(t *testing.T) {
		cueId, _, _ := m.CreateCue(clId, 50)
		req := UpdateCueAttributesRequest{CueId: cueId, Field: "label", Value: "Registered"}
		env, err := messaging.Request[UpdateCueAttributesResponse](messenger, UpdateCueAttributesRequestSubject, req)
		assert.NoError(t, err)
		assert.True(t, env.Success)
	})
}
