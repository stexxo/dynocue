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

const CueNotFound = "Cue Not Found"
const CueNumberExists = "Cue Number Already Exists"

// CreateCue

const CreateCueRequestSubject = "request.cueing.cue.create"
const CueCreatedEventSubject = "event.cueing.cue.created"

type CreateCueRequest struct {
	CueListId string  `msgpack:"cueListId" json:"cueListId"`
	CueNumber float64 `msgpack:"cueNumber" json:"cueNumber"`
}

type CreateCueResponse struct {
	CueListId string  `msgpack:"cueListId" json:"cueListId"`
	CueId     string  `msgpack:"cueId" json:"cueId"`
	CueNumber float64 `msgpack:"cueNumber" json:"cueNumber"`
}

type CueCreatedEvent struct {
	Cue types.Cue `msgpack:"cue" json:"cue"`
}

func (p *Cueing) CreateCue(sub string, req *CreateCueRequest) (*CreateCueResponse, error) {
	cue := types.Cue{
		CueListId: req.CueListId,
		CueId:     uuid.NewString(),
		Number:    req.CueNumber,
	}

	err := db.WithWrite(p.db, func(txn *memdb.Txn) error {
		if req.CueNumber == 0 {
			// Find last cue in this list
			last, err := db.GetLastTxn[types.Cue](txn, TableCues, IndexNumber, req.CueListId)
			if errors.Is(err, db.ErrItemNotFound) {
				cue.Number = 1
			} else if err != nil {
				return err
			} else {
				cue.Number = last.Number + 1
			}
		} else {
			existing, err := txn.First(TableCues, IndexNumber, req.CueListId, req.CueNumber)
			if err != nil {
				return err
			}
			if existing != nil {
				return &messaging.FriendlyError{FriendlyErr: CueNumberExists}
			}
		}

		if err := txn.Insert(TableCues, &cue); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	err = messaging.Publish(p.Messenger(), CueCreatedEventSubject, &CueCreatedEvent{
		Cue: cue,
	})
	if err != nil {
		p.Logger().Error("failed to publish cue created event", "error", err)
		return nil, err
	}

	return &CreateCueResponse{CueListId: req.CueListId, CueId: cue.CueId, CueNumber: cue.Number}, nil
}

// EnumerateCues

const EnumerateCuesRequestSubject = "request.cueing.cue.enumerate"

type EnumerateCuesRequest struct {
	CueListId string `msgpack:"cueListId" json:"cueListId" validate:"required"`
}

type EnumerateCuesResponse struct {
	Cues []types.Cue `msgpack:"cues" json:"cues"`
}

func (p *Cueing) EnumerateCues(sub string, request *EnumerateCuesRequest) (*EnumerateCuesResponse, error) {
	out, err := db.GetAllDb[types.Cue](p.db, TableCues, IndexNumber, request.CueListId)
	if err != nil {
		return nil, err
	}

	return &EnumerateCuesResponse{Cues: out}, nil
}

// GetCueByNumber

const GetCueByNumberRequestSubject = "request.cueing.cue.get.number"

type GetCueByNumberRequest struct {
	CueListNumber float64 `msgpack:"cueListNumber" json:"cueListNumber" validate:"required,gt=0"`
	CueNumber     float64 `msgpack:"cueNumber" json:"cueNumber" validate:"required,gt=0"`
}

type GetCueByNumberResponse struct {
	Cue types.Cue `msgpack:"cue" json:"cue"`
}

func (p *Cueing) GetCueByNumber(sub string, request *GetCueByNumberRequest) (*GetCueByNumberResponse, error) {
	cueList, err := db.GetFirstDb[types.CueList](p.db, TableCueLists, IndexNumber, request.CueListNumber)
	if err != nil {
		if errors.Is(err, db.ErrItemNotFound) {
			return nil, &messaging.FriendlyError{FriendlyErr: CueListNotFound}
		}
		return nil, err
	}

	cue, err := db.GetFirstDb[types.Cue](p.db, TableCues, IndexNumber, cueList.CueListId, request.CueNumber)
	if err != nil {
		if errors.Is(err, db.ErrItemNotFound) {
			return nil, &messaging.FriendlyError{FriendlyErr: CueNotFound}
		}
		return nil, err
	}

	return &GetCueByNumberResponse{Cue: *cue}, nil
}

// GetCueById

const GetCueByIdRequestSubject = "request.cueing.cue.get.id"

type GetCueByIdRequest struct {
	CueListId string `msgpack:"cueListId" json:"cueListId" validate:"required"`
	CueId     string `msgpack:"cueId" json:"cueId" validate:"required"`
}

type GetCueByIdResponse struct {
	Cue types.Cue `msgpack:"cue" json:"cue"`
}

func (p *Cueing) GetCueById(sub string, request *GetCueByIdRequest) (*GetCueByIdResponse, error) {
	cue, err := db.GetFirstDb[types.Cue](p.db, TableCues, IndexCueId, request.CueId)
	if err != nil {
		if errors.Is(err, db.ErrItemNotFound) {
			return nil, &messaging.FriendlyError{FriendlyErr: CueNotFound}
		}
		return nil, err
	}

	return &GetCueByIdResponse{Cue: *cue}, nil
}

// DeleteCue

const DeleteCueRequestSubject = "request.cueing.cue.delete"
const DeleteCueEventSubject = "event.cueing.cue.deleted"

type DeleteCueRequest struct {
	CueListId string `msgpack:"cueListId" json:"cueListId" validate:"required"`
	CueId     string `msgpack:"cueId" json:"cueId" validate:"required"`
}

type DeleteCueResponse struct{}

type CueDeletedEvent struct {
	CueListId string `msgpack:"cueListId" json:"cueListId"`
	CueId     string `msgpack:"cueId" json:"cueId"`
}

func (p *Cueing) DeleteCue(sub string, request *DeleteCueRequest) (*DeleteCueResponse, error) {
	err := db.DeleteItemFromDb[types.Cue](p.db, TableCues, IndexCueId, request.CueId)
	if err != nil {
		return nil, err
	}

	err = messaging.Publish(p.Messenger(), DeleteCueEventSubject, &CueDeletedEvent{
		CueListId: request.CueListId,
		CueId:     request.CueId,
	})
	if err != nil {
		p.Logger().Error("failed to publish cue deleted event", "error", err, "cueListId", request.CueListId, "id", request.CueId)
		return nil, err
	}

	return &DeleteCueResponse{}, nil
}

// Update Operations

// Update Events

const CueAttributesUpdatedEventSubject = "event.cueing.cue.attributes.updated"

type CueUpdatedEvent struct {
	Cue types.Cue `msgpack:"cue" json:"cue"`
}

// UpdateCueAttributes

const UpdateCueAttributesRequestSubject = "request.cueing.cue.attributes.update"

type UpdateCueAttributesRequest struct {
	CueListId string      `msgpack:"cueListId" json:"cueListId" validate:"required"`
	CueId     string      `msgpack:"cueId" json:"cueId" validate:"required"`
	Field     string      `msgpack:"field" json:"field" validate:"required"`
	Value     interface{} `msgpack:"value" json:"value"`
}

type UpdateCueAttributesResponse struct{}

func (p *Cueing) UpdateCueAttributes(sub string, request *UpdateCueAttributesRequest) (*UpdateCueAttributesResponse, error) {
	err := db.UpdateStructInDb(p.db, TableCues, IndexCueId, request.CueId, request.Field, request.Value)
	if err != nil {
		p.Logger().Error("failed to update field in cue", "error", err)
		return nil, err
	}

	cue, err := db.GetFirstDb[types.Cue](p.db, TableCues, IndexCueId, request.CueId)
	if err != nil {
		return nil, err
	}

	err = messaging.Publish(p.Messenger(), CueAttributesUpdatedEventSubject, &CueUpdatedEvent{
		Cue: *cue,
	})
	if err != nil {
		p.Logger().Error("Failed to publish updated cue", "error", err)
		return nil, err
	}

	return &UpdateCueAttributesResponse{}, nil
}
