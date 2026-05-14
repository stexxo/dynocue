// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package api

import (
	"github.com/stexxo/dynocue/components/cues/model"
	"github.com/stexxo/dynocue/util"
)

func (c *CueingApi) registerPersistenceEvents() {
	c.model.RegisterEventHandler(model.ResourceModel, model.OperationLoaded, eventHandler[PersistenceChangeEvent](c.messenger, c.logger, c.PersistenceChanged))
}

const (
	ModelLoadedEventSubject = "event.cueing.persistence.loaded"
)

type PersistenceChangeEvent struct {
}

func (c *CueingApi) PersistenceChanged(ev util.Event) (string, *PersistenceChangeEvent) {
	var sub string
	switch ev.Operation {
	case model.OperationLoaded:
		sub = ModelLoadedEventSubject
	}
	return sub, &PersistenceChangeEvent{}
}
