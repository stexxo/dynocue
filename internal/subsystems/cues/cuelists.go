package cues

import (
	"errors"

	"github.com/nats-io/nats.go"
	"gitlab.com/stexxo/dynocue/pkg/proto"
	"go.etcd.io/bbolt"
)

type CueListHandlers struct {
	m  *proto.Messenger
	db *bbolt.DB
}

// NewCueListHandlers creates and registers handlers for cue list operations.
func NewCueListHandlers(nc *nats.Conn, db *bbolt.DB) (*CueListHandlers, error) {
	clh := &CueListHandlers{
		m:  proto.NewMessenger(nc),
		db: db,
	}

	err := errors.Join(
		proto.Handle(clh.m, "show.cuelists.create", clh.CreateCueListHandler),
		proto.Handle(clh.m, "show.cuelists.list", clh.ListCueListsHandler),
		proto.Handle(clh.m, "show.cuelists.update.*", clh.UpdateCueListHandler),
		proto.Handle(clh.m, "show.cuelists.delete.*", clh.DeleteCueListHandler),
	)
	if err != nil {
		return nil, err
	}

	return clh, nil
}

type UpdateCueListRequest struct {
	Number   float32 `msgpack:"number,omitempty"`
	Label    *string `msgpack:"label,omitempty"`
	ListType *string `msgpack:"listType,omitempty" validate:"omitempty,oneof=Sequential Trigger"`
}

type UpdateCueListResponse struct {
	Number float32 `msgpack:"number,omitempty"`
}

func (clh *CueListHandlers) UpdateCueListHandler(request *UpdateCueListRequest) (proto.MessageResponse[*UpdateCueListResponse], error) {
	return proto.MessageResponse[*UpdateCueListResponse]{Body: nil}, nil
}

type DeleteCueListRequest struct{}

type DeleteCueListResponse struct {
	Success bool `msgpack:"success"`
}

func (clh *CueListHandlers) DeleteCueListHandler(request *DeleteCueListRequest) (proto.MessageResponse[*DeleteCueListResponse], error) {
	return proto.MessageResponse[*DeleteCueListResponse]{Body: nil}, nil
}

type ListCueListsRequest struct{}

type ListCueListsResponse struct {
	Lists []CueList `msgpack:"lists"`
}

type CueList struct {
	Number   float32 `msgpack:"number"`
	Label    string  `msgpack:"label"`
	ListType string  `msgpack:"listType"`
}

func (clh *CueListHandlers) ListCueListsHandler(request *ListCueListsRequest) (proto.MessageResponse[*ListCueListsResponse], error) {
	return proto.MessageResponse[*ListCueListsResponse]{Body: nil}, nil
}

type CreateCueListRequest struct {
	Number   float32 `msgpack:"number,omitempty"`
	Label    string  `msgpack:"label"`
	ListType string  `msgpack:"listType" validate:"omitempty,oneof=Sequential Trigger"`
}

type CreateCueListResponse struct {
	Number float32 `msgpack:"number,omitempty"`
}

func (clh *CueListHandlers) CreateCueListHandler(request *CreateCueListRequest) (proto.MessageResponse[*CreateCueListResponse], error) {
	return proto.MessageResponse[*CreateCueListResponse]{Body: nil}, nil
}
