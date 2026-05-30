package model

import "github.com/stexxo/dynocue/util"

const (
	ResourceFile   = "file"
	ResourceSource = "source"
	ResourceModel  = "model"

	MetadataFileId   = "FileId"
	MetadataSourceId = "SourceId"

	OperationLoaded     = "loaded"
	OperationCreated    = "created"
	OperationUpdated    = "updated"
	OperationDeleted    = "deleted"
	OperationStarted    = "started"
	OperationFinished   = "finished"
	OperationUnselected = "unselected"
)

func (m *AudioModel) RegisterEventHandler(resource, operation string, fn util.HandlerFn) {
	m.registry.Register(resource, operation, fn)
}
