// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package model

import "github.com/stexxo/dynocue/util"

const (
	ResourceCueList         = "cue_list"
	ResourceCue             = "cue"
	ResourceAction          = "action"
	ResourceActionTemplate  = "action_template"
	ResourceModel           = "cueing_model"
	ResourceCueExecution    = "cue_execution"
	ResourceActionExecution = "action_execution"

	MetadataCueListId        = "CueListId"
	MetadataCueId            = "CueId"
	MetadataActionId         = "ActionId"
	MetadataActionTemplateId = "ActionTemplateId"

	OperationLoaded     = "loaded"
	OperationCreated    = "created"
	OperationUpdated    = "updated"
	OperationDeleted    = "deleted"
	OperationStarted    = "started"
	OperationFinished   = "finished"
	OperationUnselected = "unselected"
)

func (m *CueingModel) RegisterEventHandler(resource, operation string, fn util.HandlerFn) {
	m.registry.Register(resource, operation, fn)
}
