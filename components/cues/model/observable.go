package model

import "github.com/stexxo/dynocue/util"

const (
	ResourceCueList        = "cue_list"
	ResourceCue            = "cue"
	ResourceAction         = "action"
	ResourceActionTemplate = "action_template"
	ResourceModel          = "cueing_model"

	MetadataCueListId        = "CueListId"
	MetadataCueId            = "CueId"
	MetadataActionId         = "ActionId"
	MetadataActionTemplateId = "ActionTemplateId"

	OperationLoaded  = "loaded"
	OperationCreated = "created"
	OperationUpdated = "updated"
	OperationDeleted = "deleted"
)

func (m *CueingModel) RegisterEventHandler(resource, operation string, fn util.HandlerFn) {
	m.registry.Register(resource, operation, fn)
}
