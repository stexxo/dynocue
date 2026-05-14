package api

import (
	"errors"
	"testing"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"github.com/stexxo/dynocue/components/cues/model"
	"github.com/stexxo/dynocue/components/cues/types"
	"github.com/stexxo/dynocue/core/logging"
	"github.com/stexxo/dynocue/core/messaging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testServer() (*server.Server, *nats.Conn) {
	s, _ := server.NewServer(&server.Options{DontListen: true})
	s.Start()
	nc, _ := nats.Connect("", nats.InProcessServer(s))
	return s, nc
}

func setup(t *testing.T) (*model.CueingModel, *CueingApi) {
	m, err := model.NewCueingModel()
	require.NoError(t, err)
	api := &CueingApi{
		model: m,
	}
	return m, api
}

func TestCreateCueList(t *testing.T) {
	t.Run("Success with specific number", func(t *testing.T) {
		m, api := setup(t)

		req := &CreateCueListRequest{
			Number:      10,
			CueListType: types.CueListTypeSequential,
		}
		resp, err := api.CreateCueList("test-sub", req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.NotEmpty(t, resp.CueListId)
		assert.Equal(t, uint(10), resp.Number)

		// Verify it's in the model
		cl, err := m.GetCueListById(resp.CueListId)
		assert.NoError(t, err)
		assert.Equal(t, uint(10), cl.Number)
	})

	t.Run("Success with auto-increment (number 0)", func(t *testing.T) {
		_, api := setup(t)

		req := &CreateCueListRequest{
			Number:      0,
			CueListType: types.CueListTypeSequential,
		}

		resp, err := api.CreateCueList("test-sub", req)
		assert.NoError(t, err)
		assert.Equal(t, uint(1), resp.Number)

		resp2, err := api.CreateCueList("test-sub", req)
		assert.NoError(t, err)
		assert.Equal(t, uint(2), resp2.Number)
	})

	t.Run("Error number exists", func(t *testing.T) {
		_, api := setup(t)

		req := &CreateCueListRequest{
			Number:      5,
			CueListType: types.CueListTypeSequential,
		}
		_, err := api.CreateCueList("test-sub", req)
		assert.NoError(t, err)

		// Try to create again with same number
		resp, err := api.CreateCueList("test-sub", req)
		assert.Error(t, err)
		assert.Nil(t, resp)

		assert.True(t, errors.Is(err, model.ErrNumberExists))

		var friendlyErr *messaging.FriendlyError
		assert.True(t, errors.As(err, &friendlyErr))
		assert.Equal(t, CueListNumberExists, friendlyErr.FriendlyErr)
	})
}

func TestEnumerateCueLists(t *testing.T) {
	t.Run("Success with multiple cue lists", func(t *testing.T) {
		m, api := setup(t)

		// Seed some data
		_, _, _ = m.CreateCueList(1, types.CueListTypeSequential)
		_, _, _ = m.CreateCueList(2, types.CueListTypeSequential)
		_, _, _ = m.CreateCueList(5, types.CueListTypeSequential)

		resp, err := api.EnumerateCueLists("test-sub", &EnumerateCueListsRequest{})
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp.CueLists, 3)

		// Verify numbers (should be sorted by number if using IndexNumber)
		assert.Equal(t, uint(1), resp.CueLists[0].Number)
		assert.Equal(t, uint(2), resp.CueLists[1].Number)
		assert.Equal(t, uint(5), resp.CueLists[2].Number)
	})

	t.Run("Success with no cue lists", func(t *testing.T) {
		_, api := setup(t)

		resp, err := api.EnumerateCueLists("test-sub", &EnumerateCueListsRequest{})
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Empty(t, resp.CueLists)
	})
}

func TestGetCueListByNumber(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m, api := setup(t)

		// Seed data
		_, num, err := m.CreateCueList(10, types.CueListTypeSequential)
		require.NoError(t, err)

		req := &GetCueListByNumberRequest{
			Number: float64(num),
		}

		resp, err := api.GetCueListByNumber("test-sub", req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, uint(10), resp.CueList.Number)
	})

	t.Run("Error not found", func(t *testing.T) {
		_, api := setup(t)

		req := &GetCueListByNumberRequest{
			Number: 999,
		}

		resp, err := api.GetCueListByNumber("test-sub", req)
		assert.Error(t, err)
		assert.Nil(t, resp)

		assert.True(t, errors.Is(err, model.ErrCueListNotFound))

		friendlyErr, ok := errors.AsType[*messaging.FriendlyError](err)
		assert.True(t, ok)
		assert.Equal(t, CueListNotFound, friendlyErr.FriendlyErr)
	})
}

func TestGetCueListById(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m, api := setup(t)

		// Seed data
		id, _, err := m.CreateCueList(10, types.CueListTypeSequential)
		require.NoError(t, err)

		req := &GetCueListByIdRequest{
			CueListId: id,
		}

		resp, err := api.GetCueListById("test-sub", req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, id, resp.CueList.CueListId)
		assert.Equal(t, uint(10), resp.CueList.Number)
	})

	t.Run("Error not found", func(t *testing.T) {
		_, api := setup(t)

		req := &GetCueListByIdRequest{
			CueListId: "non-existent-id",
		}

		resp, err := api.GetCueListById("test-sub", req)
		assert.Error(t, err)
		assert.Nil(t, resp)

		assert.True(t, errors.Is(err, model.ErrCueListNotFound))

		friendlyErr, ok := errors.AsType[*messaging.FriendlyError](err)
		assert.True(t, ok)
		assert.Equal(t, CueListNotFound, friendlyErr.FriendlyErr)
	})
}

func TestDeleteCueList(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m, api := setup(t)

		// Seed data
		id, _, err := m.CreateCueList(10, types.CueListTypeSequential)
		require.NoError(t, err)

		req := &DeleteCueListRequest{
			CueListId: id,
		}

		resp, err := api.DeleteCueList("test-sub", req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)

		// Verify it's gone from the model
		cl, err := m.GetCueListById(id)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, model.ErrCueListNotFound))
		assert.Nil(t, cl)

		// Idempotent Delete
		resp, err = api.DeleteCueList("test-sub", req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("Non-existent cue list", func(t *testing.T) {
		_, api := setup(t)

		req := &DeleteCueListRequest{
			CueListId: "non-existent-id",
		}

		resp, err := api.DeleteCueList("test-sub", req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})
}

func TestUpdateCueListAttributes(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m, api := setup(t)

		// Seed data
		id, _, err := m.CreateCueList(10, types.CueListTypeSequential)
		require.NoError(t, err)

		req := &UpdateCueListAttributesRequest{
			CueListId: id,
			Field:     "label",
			Value:     "New Label",
		}

		resp, err := api.UpdateCueListAttributes("test-sub", req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)

		// Verify change in model
		cl, err := m.GetCueListById(id)
		assert.NoError(t, err)
		assert.Equal(t, "New Label", cl.Label)
	})

	t.Run("Error not found", func(t *testing.T) {
		_, api := setup(t)

		req := &UpdateCueListAttributesRequest{
			CueListId: "non-existent-id",
			Field:     "label",
			Value:     "New Label",
		}

		resp, err := api.UpdateCueListAttributes("test-sub", req)
		assert.Error(t, err)
		assert.Nil(t, resp)

		assert.True(t, errors.Is(err, model.ErrCueListNotFound))

		friendlyErr, ok := errors.AsType[*messaging.FriendlyError](err)
		assert.True(t, ok)
		assert.Equal(t, CueListNotFound, friendlyErr.FriendlyErr)
	})
}

func TestRegisterCueListApis(t *testing.T) {
	s, nc := testServer()
	defer nc.Close()
	defer s.Shutdown()

	m, err := model.NewCueingModel()
	require.NoError(t, err)

	msg := messaging.NewMessenger(&messaging.MessengerCfg{
		Conn:   nc,
		Logger: logging.NewNoopLogger(),
	})

	// NewCueingApi calls registerCueListApis
	_, err = NewCueingApi(m, nil, msg, logging.NewNoopLogger())
	require.NoError(t, err)

	// Seed one cue list for testing retrieval
	id, num, err := m.CreateCueList(10, types.CueListTypeSequential)
	require.NoError(t, err)

	t.Run("CreateCueList registration", func(t *testing.T) {
		req := &CreateCueListRequest{Number: 1, CueListType: types.CueListTypeSequential}
		resp, err := messaging.Request[CreateCueListResponse](msg, CreateCueListRequestSubject, req)
		assert.NoError(t, err)
		assert.True(t, resp.Success)
	})

	t.Run("EnumerateCueLists registration", func(t *testing.T) {
		req := &EnumerateCueListsRequest{}
		resp, err := messaging.Request[EnumerateCueListsResponse](msg, EnumerateCueListsRequestSubject, req)
		assert.NoError(t, err)
		assert.True(t, resp.Success)
	})

	t.Run("GetCueListByNumber registration", func(t *testing.T) {
		req := &GetCueListByNumberRequest{Number: float64(num)}
		resp, err := messaging.Request[GetCueListByNumberResponse](msg, GetCueListByNumberRequestSubject, req)
		assert.NoError(t, err)
		assert.True(t, resp.Success)
		assert.Equal(t, num, resp.Response.CueList.Number)
	})

	t.Run("GetCueListById registration", func(t *testing.T) {
		req := &GetCueListByIdRequest{CueListId: id}
		resp, err := messaging.Request[GetCueListByIdResponse](msg, GetCueListByIdRequestSubject, req)
		assert.NoError(t, err)
		assert.True(t, resp.Success)
		assert.Equal(t, id, resp.Response.CueList.CueListId)
	})

	t.Run("DeleteCueList registration", func(t *testing.T) {
		req := &DeleteCueListRequest{CueListId: id}
		resp, err := messaging.Request[DeleteCueListResponse](msg, DeleteCueListRequestSubject, req)
		assert.NoError(t, err)
		assert.True(t, resp.Success)
	})

	t.Run("UpdateCueListAttributes registration", func(t *testing.T) {
		newId, _, _ := m.CreateCueList(20, types.CueListTypeSequential)
		req := &UpdateCueListAttributesRequest{CueListId: newId, Field: "label", Value: "New"}
		resp, err := messaging.Request[UpdateCueListAttributesResponse](msg, UpdateCueListAttributesRequestSubject, req)
		assert.NoError(t, err)
		assert.True(t, resp.Success)
	})
}
