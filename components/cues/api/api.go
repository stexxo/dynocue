// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package api

import (
	"errors"

	"github.com/stexxo/dynocue/components/cues/engine"
	"github.com/stexxo/dynocue/components/cues/model"
	"github.com/stexxo/dynocue/components/system"
	"github.com/stexxo/dynocue/core/logging"
	"github.com/stexxo/dynocue/core/messaging"
	"github.com/stexxo/dynocue/util"
)

type CueingApi struct {
	model       *model.CueingModel
	engine      *engine.CueingEngine
	persistence *system.PersistenceManager
	messenger   *messaging.Messenger
	logger      logging.Logger
}

func NewCueingApi(model *model.CueingModel, e *engine.CueingEngine, persistence *system.PersistenceManager, messaging *messaging.Messenger, logger logging.Logger) (*CueingApi, error) {
	c := &CueingApi{model: model, engine: e, persistence: persistence, messenger: messaging, logger: logger}
	err := errors.Join(
		c.registerCueListApis(),
		c.registerCueApis(),
		c.registerActionApis(),
		c.registerActionTemplateApis(),
		c.registerPersistenceApis(),
	)
	if err != nil {
		return nil, err
	}

	c.registerCueListEvents()
	c.registerCueEvents()
	c.registerActionEvents()
	c.registerActionTemplateEvents()
	c.registerPersistenceEvents()

	return c, nil
}

func eventHandler[T any](m *messaging.Messenger, l logging.Logger, evFn func(util.Event) (string, *T)) util.HandlerFn {
	return func(event util.Event) {
		sub, body := evFn(event)
		err := messaging.Publish(m, sub, body)
		if err != nil {
			l.Error("failed to publish event", "error", err, "resource", event.Resource, "operation", event.Operation)
		}
	}
}
