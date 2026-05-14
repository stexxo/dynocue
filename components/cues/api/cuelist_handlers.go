package api

import (
	"errors"

	"github.com/stexxo/dynocue/components/cues/model"
	"github.com/stexxo/dynocue/components/cues/types"
	"github.com/stexxo/dynocue/core/messaging"
)

const (
	CueListNumberExists = "Cue list number already exists"
	CueListNotFound     = "Cue list not found"
)

func (c *CueingApi) registerCueListApis() error {
	return errors.Join(
		messaging.Reply[CreateCueListRequest, CreateCueListResponse](c.messenger, true, CreateCueListRequestSubject, c.CreateCueList),
		messaging.Reply[EnumerateCueListsRequest, EnumerateCueListsResponse](c.messenger, true, EnumerateCueListsRequestSubject, c.EnumerateCueLists),
		messaging.Reply[GetCueListByNumberRequest, GetCueListByNumberResponse](c.messenger, true, GetCueListByNumberRequestSubject, c.GetCueListByNumber),
		messaging.Reply[GetCueListByIdRequest, GetCueListByIdResponse](c.messenger, true, GetCueListByIdRequestSubject, c.GetCueListById),
		messaging.Reply[DeleteCueListRequest, DeleteCueListResponse](c.messenger, true, DeleteCueListRequestSubject, c.DeleteCueList),
		messaging.Reply[UpdateCueListAttributesRequest, UpdateCueListAttributesResponse](c.messenger, true, UpdateCueListAttributesRequestSubject, c.UpdateCueListAttributes),
	)
}

const CreateCueListRequestSubject = "request.cueing.cuelists.create"

type CreateCueListRequest struct {
	Number      uint   `msgpack:"number" json:"number" validate:"gte=0"`
	CueListType string `msgpack:"cueListType" json:"cueListType" validate:"required,oneof=SEQUENTIAL"`
}

type CreateCueListResponse struct {
	CueListId string `msgpack:"cueListId" json:"cueListId"`
	Number    uint   `msgpack:"number" json:"number"`
}

func (c *CueingApi) CreateCueList(sub string, request *CreateCueListRequest) (*CreateCueListResponse, error) {
	id, num, err := c.model.CreateCueList(request.Number, request.CueListType)
	if errors.Is(err, model.ErrNumberExists) {
		return nil, errors.Join(err, &messaging.FriendlyError{FriendlyErr: CueListNumberExists})
	}
	if err != nil {
		return nil, err
	}
	return &CreateCueListResponse{CueListId: id, Number: num}, nil
}

const EnumerateCueListsRequestSubject = "request.cueing.cuelists.enumerate"

type EnumerateCueListsRequest struct{}
type EnumerateCueListsResponse struct {
	CueLists []types.CueList `msgpack:"cueLists" json:"cueLists"`
}

func (c *CueingApi) EnumerateCueLists(sub string, request *EnumerateCueListsRequest) (*EnumerateCueListsResponse, error) {
	cueLists, err := c.model.EnumerateCueLists()
	if err != nil {
		return nil, err
	}
	return &EnumerateCueListsResponse{CueLists: cueLists}, nil
}

const GetCueListByNumberRequestSubject = "request.cueing.cuelists.get.number"

type GetCueListByNumberRequest struct {
	Number float64 `msgpack:"number" json:"number" validate:"required,gt=0"`
}

type GetCueListByNumberResponse struct {
	CueList types.CueList `msgpack:"cueList" json:"cueList"`
}

func (c *CueingApi) GetCueListByNumber(sub string, request *GetCueListByNumberRequest) (*GetCueListByNumberResponse, error) {
	out, err := c.model.GetCueListByNumber(uint(request.Number))
	if errors.Is(err, model.ErrCueListNotFound) {
		return nil, errors.Join(err, &messaging.FriendlyError{FriendlyErr: CueListNotFound})
	}
	if err != nil {
		return nil, err
	}
	return &GetCueListByNumberResponse{
		CueList: *out,
	}, nil
}

const GetCueListByIdRequestSubject = "request.cueing.cuelists.get.id"

type GetCueListByIdRequest struct {
	CueListId string `msgpack:"cueListId" json:"cueListId" validate:"required"`
}

type GetCueListByIdResponse struct {
	CueList types.CueList `msgpack:"cueList" json:"cueList"`
}

func (c *CueingApi) GetCueListById(sub string, request *GetCueListByIdRequest) (*GetCueListByIdResponse, error) {
	out, err := c.model.GetCueListById(request.CueListId)
	if errors.Is(err, model.ErrCueListNotFound) {
		return nil, errors.Join(err, &messaging.FriendlyError{FriendlyErr: CueListNotFound})
	}
	if err != nil {
		return nil, err
	}

	return &GetCueListByIdResponse{
		CueList: *out,
	}, nil
}

const DeleteCueListRequestSubject = "request.cueing.cuelists.delete"

type DeleteCueListRequest struct {
	CueListId string `msgpack:"cueListId" json:"cueListId" validate:"required"`
}

type DeleteCueListResponse struct{}

func (c *CueingApi) DeleteCueList(sub string, request *DeleteCueListRequest) (*DeleteCueListResponse, error) {
	err := c.model.DeleteCueListById(request.CueListId)
	if err != nil {
		return nil, err
	}

	return &DeleteCueListResponse{}, nil
}

const UpdateCueListAttributesRequestSubject = "request.cueing.cuelists.attributes.update"

type UpdateCueListAttributesRequest struct {
	CueListId string      `msgpack:"cueListId" json:"cueListId" validate:"required"`
	Field     string      `msgpack:"field" json:"field" validate:"required,ne=id,ne=cueListType"`
	Value     interface{} `msgpack:"value" json:"value" validate:"required"`
}

type UpdateCueListAttributesResponse struct{}

func (c *CueingApi) UpdateCueListAttributes(sub string, request *UpdateCueListAttributesRequest) (*UpdateCueListAttributesResponse, error) {
	err := c.model.UpdateCueListAttribute(request.CueListId, request.Field, request.Value)
	if errors.Is(err, model.ErrCueListNotFound) {
		return nil, errors.Join(err, &messaging.FriendlyError{FriendlyErr: CueListNotFound})
	}
	if err != nil {
		return nil, err
	}

	return &UpdateCueListAttributesResponse{}, nil
}
