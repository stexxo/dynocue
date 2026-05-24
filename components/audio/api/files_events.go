// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package api

import (
	"github.com/stexxo/dynocue/components/audio/model"
	"github.com/stexxo/dynocue/util"
)

func (a *AudioAPI) registerFileEvents() {
	a.model.RegisterEventHandler(model.ResourceFile, model.OperationCreated, eventHandler[FileChangeEvent](a.messenger, a.logger, a.FileChanged))
	a.model.RegisterEventHandler(model.ResourceFile, model.OperationUpdated, eventHandler[FileChangeEvent](a.messenger, a.logger, a.FileChanged))
	a.model.RegisterEventHandler(model.ResourceFile, model.OperationDeleted, eventHandler[FileChangeEvent](a.messenger, a.logger, a.FileChanged))
}

const (
	FileCreatedEventSubject           = "event.audio.files.created"
	FileAttributesUpdatedEventSubject = "event.audio.files.updated"
	DeleteFileEventSubject            = "event.audio.files.deleted"
)

type FileChangeEvent struct {
	FileId string `msgpack:"fileId" json:"fileId"`
}

func (a *AudioAPI) FileChanged(ev util.Event) (string, *FileChangeEvent) {
	var sub string
	switch ev.Operation {
	case model.OperationUpdated:
		sub = FileAttributesUpdatedEventSubject
	case model.OperationDeleted:
		sub = DeleteFileEventSubject
	case model.OperationCreated:
		sub = FileCreatedEventSubject
	}
	return sub, &FileChangeEvent{FileId: ev.EventData[model.MetadataFileId]}
}
