// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cues

import (
	"errors"
	"fmt"

	"github.com/hashicorp/go-memdb"
	"github.com/stexxo/dynocue/components/cues/types"
	"github.com/stexxo/dynocue/core/messaging"
	"github.com/stexxo/dynocue/db"
	"github.com/stexxo/dynocue/util"
)

const ActionNotFound = "Action Not Found"
const ActionNumberExists = "Action Number Already Exists"

// CreateAction

const CreateActionRequestSubject = "request.cueing.actions.create"
const ActionCreatedEventSubject = "event.cueing.actions.created"

type CreateActionRequest struct {
	CueId        string `msgpack:"cueId" json:"cueId" validate:"required"`
	TemplateId   string `msgpack:"templateId" json:"templateId" validate:"required"`
	ActionNumber uint   `msgpack:"actionNumber" json:"actionNumber"`
}

type CreateActionResponse struct {
	Action types.Action `msgpack:"action" json:"action"`
}

type ActionCreatedEvent struct {
	Action types.Action `msgpack:"action" json:"action"`
}

func (p *Cueing) CreateAction(sub string, request *CreateActionRequest) (*CreateActionResponse, error) {
	template, err := db.GetFirstDb[types.ActionTemplate](p.runtimeDb, TableActionTemplates, IndexId, request.TemplateId)
	if err != nil {
		return nil, &messaging.FriendlyError{FriendlyErr: ActionTemplateNotFound}
	}

	cue, err := db.GetFirstDb[types.Cue](p.db, TableCues, IndexId, request.CueId)
	if err != nil {
		return nil, &messaging.FriendlyError{FriendlyErr: CueNotFound}
	}

	action := template.NewAction(cue.CueListId, request.CueId, request.ActionNumber)
	action.Number = request.ActionNumber

	err = db.WithWrite(p.db, func(txn *memdb.Txn) error {
		if action.Number == 0 {
			// Find last action in this cue
			last, err := db.GetLastTxn[types.Action](txn, TableActions, IndexNumberPrefix, request.CueId)
			if errors.Is(err, db.ErrItemNotFound) {
				action.Number = 1
			} else if err != nil {
				return err
			} else {
				action.Number = last.Number + 1
			}
		} else {
			existing, err := txn.First(TableActions, IndexNumber, request.CueId, action.Number)
			if err != nil {
				return err
			}
			if existing != nil {
				return &messaging.FriendlyError{FriendlyErr: ActionNumberExists}
			}
		}

		return txn.Insert(TableActions, action)
	})
	if err != nil {
		return nil, err
	}

	err = messaging.Publish(p.Messenger(), ActionCreatedEventSubject, &ActionCreatedEvent{
		Action: *action,
	})
	if err != nil {
		p.Logger().Error("failed to publish action created event", "error", err)
		return nil, err
	}

	return &CreateActionResponse{Action: *action}, nil
}

// EnumerateActions

const EnumerateActionsRequestSubject = "request.cueing.actions.enumerate"

type EnumerateActionsRequest struct {
	CueId string `msgpack:"cueId" json:"cueId" validate:"required"`
}

type EnumerateActionsResponse struct {
	Actions []types.Action `msgpack:"actions" json:"actions"`
}

func (p *Cueing) EnumerateActions(sub string, request *EnumerateActionsRequest) (*EnumerateActionsResponse, error) {
	out, err := db.GetAllDb[types.Action](p.db, TableActions, IndexNumberPrefix, request.CueId)
	if err != nil {
		return nil, err
	}

	return &EnumerateActionsResponse{Actions: out}, nil
}

// GetActionById

const GetActionByIdRequestSubject = "request.cueing.actions.get.id"

type GetActionByIdRequest struct {
	ActionId string `msgpack:"actionId" json:"actionId" validate:"required"`
}

type GetActionByIdResponse struct {
	Action types.Action `msgpack:"action" json:"action"`
}

func (p *Cueing) GetActionById(sub string, request *GetActionByIdRequest) (*GetActionByIdResponse, error) {
	action, err := db.GetFirstDb[types.Action](p.db, TableActions, IndexId, request.ActionId)
	if err != nil {
		if errors.Is(err, db.ErrItemNotFound) {
			return nil, &messaging.FriendlyError{FriendlyErr: ActionNotFound}
		}
		return nil, err
	}

	return &GetActionByIdResponse{Action: *action}, nil
}

// DeleteAction

const DeleteActionRequestSubject = "request.cueing.actions.delete"
const ActionDeletedEventSubject = "event.cueing.actions.deleted"

type DeleteActionRequest struct {
	ActionId string `msgpack:"actionId" json:"actionId" validate:"required"`
}

type DeleteActionResponse struct{}

type ActionDeletedEvent struct {
	CueListId string `msgpack:"cueListId" json:"cueListId"`
	CueId     string `msgpack:"cueId" json:"cueId"`
	ActionId  string `msgpack:"actionId" json:"actionId"`
}

func (p *Cueing) DeleteAction(sub string, request *DeleteActionRequest) (*DeleteActionResponse, error) {
	action, err := db.GetFirstDb[types.Action](p.db, TableActions, IndexId, request.ActionId)
	if err != nil {
		if errors.Is(err, db.ErrItemNotFound) {
			return nil, &messaging.FriendlyError{FriendlyErr: ActionNotFound}
		}
		return nil, err
	}

	err = db.DeleteItemFromDb[types.Action](p.db, TableActions, IndexId, request.ActionId)
	if err != nil {
		return nil, err
	}

	err = messaging.Publish(p.Messenger(), ActionDeletedEventSubject, &ActionDeletedEvent{
		CueListId: action.CueListId,
		CueId:     action.CueId,
		ActionId:  request.ActionId,
	})
	if err != nil {
		p.Logger().Error("failed to publish action deleted event", "error", err, "actionId", request.ActionId)
		return nil, err
	}

	return &DeleteActionResponse{}, nil
}

// UpdateAction

const UpdateActionRequestSubject = "request.cueing.actions.update"
const ActionUpdatedEventSubject = "event.cueing.actions.updated"

type UpdateActionRequest struct {
	ActionId string      `msgpack:"actionId" json:"actionId" validate:"required"`
	Field    string      `msgpack:"field" json:"field" validate:"required,ne=id,ne=subject,ne=templateId,ne=fields"`
	Value    interface{} `msgpack:"value" json:"value"`
}

type UpdateActionResponse struct{}

type ActionUpdatedEvent struct {
	CueListId string `msgpack:"cueListId" json:"cueListId"`
	CueId     string `msgpack:"cueId" json:"cueId"`
	ActionId  string `msgpack:"actionId" json:"actionId"`
	Field     string `msgpack:"field" json:"field"`
}

func (p *Cueing) UpdateAction(sub string, request *UpdateActionRequest) (*UpdateActionResponse, error) {
	err := db.UpdateStructInDb[types.Action](p.db, TableActions, IndexId, request.ActionId, request.Field, request.Value)
	if err != nil {
		p.Logger().Error("failed to update field in action", "error", err)
		return nil, err
	}

	action, err := db.GetFirstDb[types.Action](p.db, TableActions, IndexId, request.ActionId)
	if err != nil {
		p.Logger().Error("failed to get action for event", "error", err)
		return nil, err
	}

	err = messaging.Publish(p.Messenger(), ActionUpdatedEventSubject, &ActionUpdatedEvent{
		CueListId: action.CueListId,
		CueId:     action.CueId,
		ActionId:  request.ActionId,
		Field:     request.Field,
	})
	if err != nil {
		p.Logger().Error("failed to publish updated action", "error", err)
		return nil, err
	}

	return &UpdateActionResponse{}, nil
}

// UpdateActionField

const UpdateActionFieldRequestSubject = "request.cueing.actions.field.update"

type UpdateActionFieldRequest struct {
	ActionId  string      `msgpack:"actionId" json:"actionId" validate:"required"`
	FieldName string      `msgpack:"fieldName" json:"fieldName" validate:"required"`
	Value     interface{} `msgpack:"value" json:"value"`
}

type UpdateActionFieldResponse struct{}

func (p *Cueing) UpdateActionField(sub string, request *UpdateActionFieldRequest) (*UpdateActionFieldResponse, error) {
	var action types.Action
	err := db.WithWrite(p.db, func(txn *memdb.Txn) error {
		original, err := db.GetFirstTxn[types.Action](txn, TableActions, IndexId, request.ActionId)
		if err != nil {
			return err
		}

		// Deep copy of the top level struct and nested slices
		action = *util.DeepCopyStruct(original)

		foundField := false
		for i := range action.Fields {
			if action.Fields[i].FieldName == request.FieldName {
				action.Fields[i].Value = request.Value
				foundField = true
				break
			}
		}

		if !foundField {
			return fmt.Errorf("field %s not found in action %s", request.FieldName, request.ActionId)
		}

		return txn.Insert(TableActions, &action)
	})
	if err != nil {
		p.Logger().Error("failed to update action field", "error", err)
		return nil, err
	}

	err = messaging.Publish(p.Messenger(), ActionUpdatedEventSubject, &ActionUpdatedEvent{
		CueListId: action.CueListId,
		CueId:     action.CueId,
		ActionId:  request.ActionId,
	})
	if err != nil {
		p.Logger().Error("failed to publish updated action", "error", err)
		return nil, err
	}

	return &UpdateActionFieldResponse{}, nil
}
