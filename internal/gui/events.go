// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package gui

import (
	"errors"
	"log/slog"

	"github.com/stexxo/dynocue/api/cues"
	apibus "github.com/stexxo/dynocue/pkg/bus"
)

func (c *Commands) SubscribeToAll() error {
	return errors.Join(
		subscribe(c, cues.EventNewCueList, emitEventHandler[*cues.NewCueListEvent](c)),
		subscribe(c, cues.EventUpdateCueList, emitEventHandler[*cues.UpdateCueListEvent](c)),
		subscribe(c, cues.EventDeleteCueList, emitEventHandler[*cues.DeleteCueListEvent](c)),

		subscribe(c, cues.EventNewCue, emitEventHandler[*cues.NewCueEvent](c)),
		subscribe(c, cues.EventUpdateCue, emitEventHandler[*cues.UpdateCueEvent](c)),
		subscribe(c, cues.EventDeleteCue, emitEventHandler[*cues.DeleteCueEvent](c)),

		subscribe(c, cues.EventNewAction, emitEventHandler[*cues.NewActionEvent](c)),
		subscribe(c, cues.EventUpdateAction, emitEventHandler[*cues.UpdateActionEvent](c)),
		subscribe(c, cues.EventDeleteAction, emitEventHandler[*cues.DeleteActionEvent](c)),
	)
}

func subscribe[T any](c *Commands, subject string, handler apibus.SubscribeHandler[*T]) error {
	sub, err := apibus.Subscribe[T](c.conn, subject, handler)
	if err != nil {
		return err
	}
	c.subscriptions = append(c.subscriptions, sub)
	return nil
}

func emitEventHandler[T any](c *Commands) apibus.SubscribeHandler[T] {
	return func(s string, t T) {
		slog.Debug("got event " + s)
		c.app.Event.Emit(s, t)
	}
}
