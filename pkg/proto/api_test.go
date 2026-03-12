package proto

import (
	"fmt"
	"testing"
	"time"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vmihailenco/msgpack/v5"
)

// RunServer starts a local NATS server for testing.
func RunServer(t *testing.T) *server.Server {
	t.Helper()
	opts := &server.Options{
		Port: -1, // Random port
	}
	s, err := server.NewServer(opts)
	require.NoError(t, err)
	go s.Start()
	if !s.ReadyForConnections(5 * time.Second) {
		t.Fatal("NATS server not ready")
	}
	return s
}

func TestErrorCodes(t *testing.T) {
	assert.Equal(t, 1000, ErrCodeSystemError)
	assert.Equal(t, 1001, ErrCodeInvalidPayload)
	assert.Equal(t, 1002, ErrCodeNotFound)
	assert.Equal(t, 1003, ErrCodeTimeout)
}

func TestPublishSubscribe(t *testing.T) {
	s := RunServer(t)
	defer s.Shutdown()

	nc, err := nats.Connect(s.ClientURL())
	require.NoError(t, err)
	defer nc.Close()

	type TestMsg struct {
		Name string `msgpack:"name"`
		Age  int    `msgpack:"age"`
	}

	subject := "test.pubsub"
	received := make(chan TestMsg, 1)

	sub, err := Subscribe(nc, subject, func(m TestMsg) {
		received <- m
	})
	require.NoError(t, err)
	defer sub.Unsubscribe()

	sentMsg := TestMsg{Name: "Alice", Age: 30}
	err = Publish(nc, subject, sentMsg)
	require.NoError(t, err)

	select {
	case msg := <-received:
		assert.Equal(t, sentMsg, msg)
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for message")
	}
}

func TestRequestRespond_Success(t *testing.T) {
	s := RunServer(t)
	defer s.Shutdown()

	nc, err := nats.Connect(s.ClientURL())
	require.NoError(t, err)
	defer nc.Close()

	type ReqData struct {
		ID int `msgpack:"id"`
	}
	type ResData struct {
		Message string `msgpack:"message"`
	}

	subject := "test.request"
	sub, err := Respond(nc, subject, func(req *ReqData) (MessageResponse[*ResData], error) {
		res := &ResData{Message: "Handled"}
		return MessageResponse[*ResData]{Body: &res}, nil
	})
	require.NoError(t, err)
	defer sub.Unsubscribe()

	req := ReqData{ID: 1}
	resp, msgErr, err := Request[ResData](nc, subject, req, time.Second)
	require.NoError(t, err)
	assert.Nil(t, msgErr)
	require.NotNil(t, resp)
	assert.Equal(t, "Handled", resp.Message)
}

func TestRequestRespond_Error(t *testing.T) {
	s := RunServer(t)
	defer s.Shutdown()

	nc, err := nats.Connect(s.ClientURL())
	require.NoError(t, err)
	defer nc.Close()

	type Req struct{}
	type Res struct{}

	subject := "test.error"
	expectedMsgErr := &MsgError{
		Code:    ErrCodeNotFound,
		Message: "Not found",
	}

	sub, err := Respond(nc, subject, func(req *Req) (MessageResponse[*Res], error) {
		return MessageResponse[*Res]{Body: nil, Error: expectedMsgErr}, nil
	})
	require.NoError(t, err)
	defer sub.Unsubscribe()

	resp, msgErr, err := Request[Res](nc, subject, Req{}, time.Second)
	require.NotNil(t, msgErr)
	assert.Nil(t, resp)
	assert.Equal(t, expectedMsgErr.Code, msgErr.Code)
	assert.Equal(t, expectedMsgErr.Message, msgErr.Message)
}

func TestRequest_Timeout(t *testing.T) {
	s := RunServer(t)
	defer s.Shutdown()

	nc, err := nats.Connect(s.ClientURL())
	require.NoError(t, err)
	defer nc.Close()

	type Req struct{}
	type Res struct{}

	subject := "test.timeout"
	// No one is responding on this subject

	resp, msgErr, err := Request[Res](nc, subject, Req{}, 100*time.Millisecond)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "nats request failed")
	assert.Nil(t, msgErr)
	assert.Nil(t, resp)
}

func TestRespond_MarshalError(t *testing.T) {
	s := RunServer(t)
	defer s.Shutdown()

	nc, err := nats.Connect(s.ClientURL())
	require.NoError(t, err)
	defer nc.Close()

	type Req struct{}
	type Res struct {
		Channel chan int // Channels cannot be marshaled by msgpack
	}

	subject := "test.marshal_error"
	sub, err := Respond(nc, subject, func(req *Req) (MessageResponse[*Res], error) {
		res := &Res{Channel: make(chan int)}
		return MessageResponse[*Res]{Body: &res}, nil
	})
	require.NoError(t, err)
	defer sub.Unsubscribe()

	resp, msgErr, err := Request[Res](nc, subject, Req{}, time.Second)
	require.NoError(t, err)
	require.NotNil(t, msgErr)
	assert.Nil(t, resp)
	assert.Equal(t, ErrCodeSystemError, msgErr.Code)
	assert.Contains(t, msgErr.Message, "System error during marshaling")
}

func TestRespond_InvalidPayload(t *testing.T) {
	s := RunServer(t)
	defer s.Shutdown()

	nc, err := nats.Connect(s.ClientURL())
	require.NoError(t, err)
	defer nc.Close()

	type Req struct {
		Name string `msgpack:"name"`
	}
	type Res struct{}

	subject := "test.invalid_payload"
	sub, err := Respond(nc, subject, func(req *Req) (MessageResponse[*Res], error) {
		return MessageResponse[*Res]{}, nil
	})
	require.NoError(t, err)
	defer sub.Unsubscribe()

	// Send raw data that is not a valid msgpack for the expected type
	// 0xc1 is "never used" in msgpack spec
	invalidData := []byte{0xc1}

	respMsg, err := nc.Request(subject, invalidData, time.Second)
	require.NoError(t, err)

	var msgResp msgResponse
	err = msgpack.Unmarshal(respMsg.Data, &msgResp)
	require.NoError(t, err)

	require.NotNil(t, msgResp.Error)
	assert.Equal(t, ErrCodeInvalidPayload, msgResp.Error.Code)
	assert.Contains(t, msgResp.Error.Message, "Invalid request payload")
}

func TestRequest_ErrorWithBody(t *testing.T) {
	s := RunServer(t)
	defer s.Shutdown()

	nc, err := nats.Connect(s.ClientURL())
	require.NoError(t, err)
	defer nc.Close()

	type Req struct{}
	type Res struct {
		Name string `msgpack:"name"`
	}

	subject := "test.error_with_body"
	expectedMsgErr := &MsgError{
		Code:    ErrCodeNotFound,
		Message: "Not found but here is some data",
	}
	expectedRes := Res{Name: "Partial Data"}

	sub, err := Respond(nc, subject, func(req *Req) (MessageResponse[*Res], error) {
		res := &expectedRes
		return MessageResponse[*Res]{
			Body:  &res,
			Error: expectedMsgErr,
		}, nil
	})
	require.NoError(t, err)
	defer sub.Unsubscribe()

	resp, msgErr, err := Request[Res](nc, subject, Req{}, time.Second)
	require.NoError(t, err)
	require.NotNil(t, msgErr)
	require.NotNil(t, resp)
	assert.Equal(t, expectedMsgErr.Code, msgErr.Code)
	assert.Equal(t, expectedRes.Name, resp.Name)
}

func TestRespond_HandlerError(t *testing.T) {
	s := RunServer(t)
	defer s.Shutdown()

	nc, err := nats.Connect(s.ClientURL())
	require.NoError(t, err)
	defer nc.Close()

	type Req struct{}
	type Res struct{}

	subject := "test.handler_error"
	sub, err := Respond(nc, subject, func(req *Req) (MessageResponse[*Res], error) {
		return MessageResponse[*Res]{}, fmt.Errorf("something went wrong")
	})
	require.NoError(t, err)
	defer sub.Unsubscribe()

	resp, msgErr, err := Request[Res](nc, subject, Req{}, time.Second)
	require.NoError(t, err)
	require.NotNil(t, msgErr)
	assert.Nil(t, resp)
	assert.Equal(t, ErrCodeSystemError, msgErr.Code)
	assert.Equal(t, "something went wrong", msgErr.Message)
}

func TestMessenger(t *testing.T) {
	s := RunServer(t)
	defer s.Shutdown()

	nc, err := nats.Connect(s.ClientURL())
	require.NoError(t, err)
	defer nc.Close()

	m := NewMessenger(nc)
	defer m.Close()

	type Req struct{ Name string }
	type Res struct{ Greeting string }

	err = Handle(m, "greet", func(r *Req) (MessageResponse[*Res], error) {
		res := &Res{Greeting: "Hello " + r.Name}
		return MessageResponse[*Res]{Body: &res}, nil
	})
	require.NoError(t, err)
	assert.Len(t, m.Subscriptions(), 1)

	resp, msgErr, err := Request[Res](nc, "greet", Req{Name: "World"}, time.Second)
	require.NoError(t, err)
	assert.Nil(t, msgErr)
	require.NotNil(t, resp)
	assert.Equal(t, "Hello World", resp.Greeting)

	// Test error propagation (Handle now returns error directly)
	err = Handle(m, "invalid.subject", func(r *Req) (MessageResponse[*Res], error) {
		return MessageResponse[*Res]{}, nil
	})
	require.NoError(t, err)
	assert.Len(t, m.Subscriptions(), 2)

	// Verify map storage by subject
	assert.Len(t, m.subs["greet"], 1)
	assert.Len(t, m.subs["invalid.subject"], 1)
}

func TestRespond_ZeroStructErrorBody(t *testing.T) {
	s := RunServer(t)
	defer s.Shutdown()

	nc, err := nats.Connect(s.ClientURL())
	require.NoError(t, err)
	defer nc.Close()

	type Req struct{}
	type Res struct {
		ID int `msgpack:"id"`
	}

	subject := "test.zero_struct_error"
	sub, err := Respond(nc, subject, func(req *Req) (MessageResponse[*Res], error) {
		// Return an error but an empty (zero-value) Res struct
		return MessageResponse[*Res]{
			Error: &MsgError{Code: ErrCodeNotFound, Message: "Not found"},
		}, nil
	})
	require.NoError(t, err)
	defer sub.Unsubscribe()

	resp, msgErr, err := Request[Res](nc, subject, Req{}, time.Second)
	require.NoError(t, err)
	require.NotNil(t, msgErr)
	// Currently, this will NOT be nil because Res is a struct and will be marshaled as {id:0}
	assert.Nil(t, resp, "Expected nil body when error is present and body is zero-value struct")
}
