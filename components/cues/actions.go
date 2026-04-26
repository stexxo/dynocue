package cues

import (
	"github.com/stexxo/dynocue/components/cues/types"
	"github.com/stexxo/dynocue/core/messaging"
)

const ActionNotFound = "Action not found"
const ActionNumberExists = "Action number already exists"

const CreateActionRequestSubject = "request.cueing.actions.create"
const ActionCreatedEventSubject = "request.cueing.actions.created"

type CreateActionRequest struct {
	CueListId string `msgpack:"cueListId" json:"cueListId"`
	CueId     string `msgpack:"cueId" json:"cueId"`
	Number    int    `msgpack:"number" json:"number"`

	TemplateId string `msgpack:"templateId" json:"templateId"`
}

type CreateActionResponse struct{}

type ActionCreatedEvent struct {
	CueListId string `msgpack:"cueListId" json:"cueListId"`
	CueId     string `msgpack:"cueId" json:"cueId"`
	ActionId  string `msgpack:"actionId" json:"actionId"`
}

func (p *Cueing) CreateAction(sub string, req CreateActionRequest) (*CreateActionResponse, error) {

	var action *types.Action

	if req.TemplateId != "" {
		actionTemplate := p.actionTemplates.GetTemplateById(req.TemplateId)
		if actionTemplate == nil {
			return nil, &messaging.FriendlyError{FriendlyErr: ActionTemplateNotFound}
		}
		action = types.NewActionByTemplate(actionTemplate)
	} else {
		action = types.NewAction()
	}

	cl, err := p.getCueListById(req.CueListId)
	if err != nil {
		return nil, &messaging.FriendlyError{FriendlyErr: CueListNotFound}
	}

	c, err := p.getCueById(cl, req.CueId)
	if err != nil {
		return nil, &messaging.FriendlyError{FriendlyErr: CueNotFound}
	}

	ok := c.Actions.Add(action)
	if !ok {
		return nil, &messaging.FriendlyError{FriendlyErr: ActionNumberExists}
	}

	return &CreateActionResponse{}, nil
}
