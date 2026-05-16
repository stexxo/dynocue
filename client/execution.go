// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package client

import (
	"errors"
	"fmt"

	"github.com/stexxo/dynocue/components/cues/api"
	"github.com/stexxo/dynocue/core/messaging"
)

var ErrNoNextCue = errors.New("no next cue found")
var ErrNoCueSelected = errors.New("no cue selected")

func (c *Client) GoToCue(cueId string) error {
	resp, err := messaging.Request[api.GoToCueResponse](c.messenger, api.GoToCueRequestSubject, &api.GoToCueRequest{
		CueId: cueId,
	})
	if err != nil {
		return err
	}

	if resp.Success {
		return nil
	}

	if resp.Error == api.CueNotFound {
		return ErrCueNotFound
	}

	if resp.Error == api.CueListNotFound {
		return ErrCueListNotFound
	}

	return fmt.Errorf("failed to go to cue: %s", resp.Error)
}

func (c *Client) GoToNextCue(cueListId string) error {
	resp, err := messaging.Request[api.GoToNextCueResponse](c.messenger, api.GoToNextCueRequestSubject, &api.GoToNextCueRequest{
		CueListId: cueListId,
	})
	if err != nil {
		return err
	}

	if resp.Success {
		return nil
	}

	if resp.Error == api.CueListNotFound {
		return ErrCueListNotFound
	}

	if resp.Error == api.NoNextCue {
		return ErrNoNextCue
	}

	if resp.Error == api.NoCueSelected {
		return ErrNoCueSelected
	}

	return fmt.Errorf("failed to go to next cue: %s", resp.Error)
}

func (c *Client) OnExecutionStarted(handler EventCallback[api.ExecutionChangeEvent]) error {
	err := messaging.Subscribe[api.ExecutionChangeEvent](c.messenger, false, api.ExecutionStartedEventSubject, func(s string, e *api.ExecutionChangeEvent) {
		handler(s, e)
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe to execution started events: %w", err)
	}
	return nil
}

func (c *Client) OnExecutionFinished(handler EventCallback[api.ExecutionChangeEvent]) error {
	err := messaging.Subscribe[api.ExecutionChangeEvent](c.messenger, false, api.ExecutionFinishedEventSubject, func(s string, e *api.ExecutionChangeEvent) {
		handler(s, e)
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe to execution finished events: %w", err)
	}
	return nil
}

func (c *Client) OnExecutionUnselected(handler EventCallback[api.ExecutionChangeEvent]) error {
	err := messaging.Subscribe[api.ExecutionChangeEvent](c.messenger, false, api.ExecutionUnselectedEventSubject, func(s string, e *api.ExecutionChangeEvent) {
		handler(s, e)
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe to execution unselected events: %w", err)
	}
	return nil
}

func (c *Client) OnExecutionDeleted(handler EventCallback[api.ExecutionChangeEvent]) error {
	err := messaging.Subscribe[api.ExecutionChangeEvent](c.messenger, false, api.ExecutionDeletedEventSubject, func(s string, e *api.ExecutionChangeEvent) {
		handler(s, e)
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe to execution deleted events: %w", err)
	}
	return nil
}
