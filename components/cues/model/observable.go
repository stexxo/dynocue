package model

import "github.com/stexxo/dynocue/util"

const (
	ResourceCueList        = "cue_list"
	ResourceCue            = "cue"
	ResourceAction         = "action"
	ResourceActionTemplate = "action_template"

	OperationCreated = "created"
	OperationUpdated = "updated"
	OperationDeleted = "deleted"
)

func (m *CueingModel) RegisterEventHandler(resource, action string, fn util.HandlerFn) {
	m.registry.Register(resource, action, fn)
}
