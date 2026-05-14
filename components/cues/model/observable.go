package model

import "github.com/stexxo/dynocue/util"

const (
	ResourceCueList        = "cue_list"
	ResourceCue            = "cue"
	ResourceAction         = "action"
	ResourceActionTemplate = "action_template"

	MetadataCueListId        = "CueListId"
	MetadataCueId            = "CueId"
	MetadataActionId         = "ActionId"
	MetadataActionTemplateId = "ActionTemplateId"

	OperationCreated = "created"
	OperationUpdated = "updated"
	OperationDeleted = "deleted"
)

func (m *CueingModel) RegisterEventHandler(resource, operation string, fn util.HandlerFn) {
	m.registry.Register(resource, operation, fn)
}
