// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cues

import (
	"errors"

	"github.com/google/uuid"
	"github.com/hashicorp/go-memdb"
	"github.com/stexxo/dynocue/components/cues/types"
	"github.com/stexxo/dynocue/core/messaging"
	"github.com/stexxo/dynocue/db"
)

const CueListNumberExists = "Cue List Number Already Exists"
const CueListNotFound = "Cue List Not Found."

// CreateCueList

const CreateCueListRequestSubject = "request.cueing.cuelists.create"
const CueListCreatedEventSubject = "event.cueing.cuelists.created"

type CreateCueListRequest struct {
	Number      int    `msgpack:"number" json:"number" validate:"gte=0"`
	CueListType string `msgpack:"cueListType" json:"cueListType" validate:"required,oneof=SEQUENTIAL"`
}

type CreateCueListResponse struct {
	Id     string `msgpack:"id" json:"id"`
	Number int    `msgpack:"number" json:"number"`
}

type CueListCreatedEvent struct {
	CueListId string `msgpack:"cueListId" json:"cueListId"`
}

func (p *Cueing) CreateCueList(sub string, request *CreateCueListRequest) (*CreateCueListResponse, error) {
	cl := types.CueList{
		CueListId:   uuid.NewString(),
		Number:      request.Number,
		CueListType: request.CueListType,
	}

	err := db.WithWrite(p.db, func(txn *memdb.Txn) error {
		if request.Number == 0 {
			last, err := db.GetLastTxn[types.CueList](txn, TableCueLists, IndexNumber)
			if errors.Is(err, db.ErrItemNotFound) {
				cl.Number = 1
			} else if err != nil {
				return err
			} else {
				cl.Number = last.Number + 1
			}
		} else {
			existing, err := txn.First("cuelist", "number", request.Number)
			if err != nil {
				return err
			}
			if existing != nil {
				return &messaging.FriendlyError{FriendlyErr: CueListNumberExists}
			}
		}

		if err := txn.Insert("cuelist", cl); err != nil {
			return err
		}

		txn.Commit()
		return nil
	})
	if err != nil {
		return nil, err
	}

	err = messaging.Publish(p.Messenger(), CueListCreatedEventSubject, &CueListCreatedEvent{CueListId: cl.CueListId})
	if err != nil {
		p.Logger().Error("Failed to publish cue list created event", "error", err, "cueListNumber", request.Number)
		return nil, err
	}

	return &CreateCueListResponse{
		Id:     cl.CueListId,
		Number: cl.Number,
	}, nil

}

// EnumerateCueLists

const EnumerateCueListsRequestSubject = "request.cueing.cuelists.enumerate"

type EnumerateCueListsRequest struct{}

type EnumerateCueListsResponse struct {
	CueLists []types.CueList `msgpack:"cueLists" json:"cueLists"`
}

type CueListEnumeration struct {
	Number      float64 `msgpack:"number" json:"number"`
	Label       string  `msgpack:"label" json:"label"`
	CueListType string  `msgpack:"cueListType" json:"cueListType"`
}

func (p *Cueing) EnumerateCueLists(sub string, request *EnumerateCueListsRequest) (*EnumerateCueListsResponse, error) {
	out, err := db.GetAllDb[types.CueList](p.db, TableCueLists, IndexNumber)
	if err != nil {
		return nil, err
	}
	return &EnumerateCueListsResponse{CueLists: out}, nil
}

// GetCueListByNumber

const GetCueListByNumberRequestSubject = "request.cueing.cuelists.get.number"

type GetCueListByNumberRequest struct {
	Number float64 `msgpack:"number" json:"number" validate:"required,gt=0"`
}

type GetCueListByNumberResponse struct {
	CueList types.CueList `msgpack:"cueList" json:"cueList"`
}

func (p *Cueing) GetCueListByNumber(sub string, request *GetCueListByNumberRequest) (*GetCueListByNumberResponse, error) {
	out, err := db.GetFirstDb[types.CueList](p.db, TableCueLists, "number", request.Number)
	if err != nil {
		return nil, err
	}
	return &GetCueListByNumberResponse{
		CueList: *out,
	}, nil
}

// GetCueListById

const GetCueListByIdRequestSubject = "request.cueing.cuelists.get.id"

type GetCueListByIdRequest struct {
	Id string `msgpack:"id" json:"id" validate:"required"`
}

type GetCueListByIdResponse struct {
	CueList types.CueList `msgpack:"cueList" json:"cueList"`
}

func (p *Cueing) GetCueListById(sub string, request *GetCueListByIdRequest) (*GetCueListByIdResponse, error) {
	out, err := db.GetFirstDb[types.CueList](p.db, TableCueLists, IndexCueListId, request.Id)
	if err != nil {
		return nil, err
	}

	return &GetCueListByIdResponse{
		CueList: *out,
	}, nil
}

// RenumberCueList

const RenumberCueListRequestSubject = "request.cueing.cuelists.renumber"
const RenumberCueListEventSubject = "event.cueing.cuelists.renumber"

type RenumberCueListsRequest struct {
	Id        string  `msgpack:"id" json:"id" validate:"required"`
	NewNumber float64 `msgpack:"newNumber" json:"newNumber" validate:"required,gt=0"`
}

// DeleteCueList

const DeleteCueListRequestSubject = "request.cueing.cuelists.delete"
const DeleteCueListEventSubject = "event.cueing.cuelists.deleted"

type DeleteCueListsRequest struct {
	Id string `msgpack:"id" json:"id" validate:"required"`
}

type DeleteCueListsResponse struct{}

type CueListDeletedEvent struct {
	Id string `msgpack:"id" json:"id"`
}

func (p *Cueing) DeleteCueList(sub string, request *DeleteCueListsRequest) (*DeleteCueListsResponse, error) {
	err := db.DeleteItemFromDb[types.CueList](p.db, TableCueLists, "id", request.Id)
	if err != nil {
		return nil, err
	}

	err = messaging.Publish(p.Messenger(), DeleteCueListEventSubject, &CueListDeletedEvent{
		Id: request.Id,
	})

	if err != nil {
		return nil, err
	}

	return &DeleteCueListsResponse{}, nil
}

// Update Operations

// Update Events

const CueListAttributesUpdatedEventSubject = "event.cueing.cuelists.attributes.updated"

type CueListAttributesUpdatedEvent struct {
	CueListId string `msgpack:"cueListId" json:"cueListId"`
}

// UpdateCueListAttributes

const UpdateCueListAttributesRequestSubject = "request.cueing.cuelists.attributes.update"

type UpdateCueListAttributesRequest struct {
	Id    string      `msgpack:"id" json:"id" validate:"required"`
	Field string      `msgpack:"field" json:"field" validate:"required,ne=id,ne=number,ne=cueListType"`
	Value interface{} `msgpack:"value" json:"value" validate:"required"`
}

type UpdateCueListAttributesResponse struct{}

func (p *Cueing) UpdateCueListAttributes(sub string, request *UpdateCueListAttributesRequest) (*UpdateCueListAttributesResponse, error) {
	err := db.UpdateStructInDb(p.db, TableCueLists, IndexCueListId, request.Id, request.Field, request.Value)
	if err != nil {
		return nil, err
	}

	err = messaging.Publish(p.Messenger(), CueListAttributesUpdatedEventSubject, &CueListAttributesUpdatedEvent{
		CueListId: request.Id,
	})
	if err != nil {
		p.Logger().Error("Failed to publish updated cue list attributes", "error", err)
		return nil, err
	}

	return &UpdateCueListAttributesResponse{}, nil
}
