// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package api

import (
	"github.com/stexxo/dynocue/components/cues/model"
	"github.com/stexxo/dynocue/util"
)

func (a *AudioAPI) registerPersistenceEvents() {
	a.model.RegisterEventHandler(model.ResourceModel, model.OperationLoaded, eventHandler[PersistenceChangeEvent](a.messenger, a.logger, a.PersistenceChanged))
}

const (
	ModelLoadedEventSubject = "event.audio.persistence.loaded"
)

type PersistenceChangeEvent struct {
}

func (a *AudioAPI) PersistenceChanged(ev util.Event) (string, *PersistenceChangeEvent) {
	var sub string
	switch ev.Operation {
	case model.OperationLoaded:
		sub = ModelLoadedEventSubject
	}
	return sub, &PersistenceChangeEvent{}
}
