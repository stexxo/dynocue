package gui

import (
	"errors"

	apibus "gitlab.com/stexxo/dynocue/api/bus"
	"gitlab.com/stexxo/dynocue/api/cues"
)

func (c *Commands) SubscribeToAll() error {
	return errors.Join(
		subscribe(c, cues.EventNewCueList, emitEventHandler[cues.NewCueListEvent](c)),
		subscribe(c, cues.EventUpdateCueList, emitEventHandler[cues.UpdateCueListMetadataEvent](c)),
		subscribe(c, cues.EventDeleteCueList, emitEventHandler[cues.DeleteCueListEvent](c)),
	)
}

func subscribe[T any](c *Commands, subject string, handler apibus.SubscribeHandler[T]) error {
	sub, err := apibus.Subscribe[T](c.conn, subject, handler)
	if err != nil {
		return err
	}
	c.subscriptions = append(c.subscriptions, sub)
	return nil
}

func emitEventHandler[T any](c *Commands) apibus.SubscribeHandler[T] {
	return func(s string, t T) {
		c.app.Event.Emit(s, t)
	}
}
