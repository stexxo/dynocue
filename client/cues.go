// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package client

import (
	"errors"
	"fmt"

	"github.com/stexxo/dynocue/components/cues"
	"github.com/stexxo/dynocue/components/cues/types"
	"github.com/stexxo/dynocue/core/messaging"
)

var ErrCueExists = errors.New("cue with provided number already exists")
var ErrCueNotFound = errors.New("cue not found")

func (c *Client) CreateCue(cueListId string, cueNumber uint) (uint, error) {
	resp, err := messaging.Request[cues.CreateCueResponse](c.messenger, cues.CreateCueRequestSubject, &cues.CreateCueRequest{
		CueListId: cueListId,
		CueNumber: cueNumber,
	})
	if err != nil {
		return 0, err
	}
	if resp.Success {
		return resp.Response.CueNumber, nil
	}

	if resp.Error == cues.CueNumberExists {
		return 0, ErrCueExists
	}

	return 0, fmt.Errorf("failed to create cue: %s", resp.Error)
}

func (c *Client) EnumerateCues(cueListId string) ([]types.Cue, error) {
	resp, err := messaging.Request[cues.EnumerateCuesResponse](c.messenger, cues.EnumerateCuesRequestSubject, &cues.EnumerateCuesRequest{
		CueListId: cueListId,
	})
	if err != nil {
		return nil, err
	}

	if resp.Success {
		return resp.Response.Cues, nil
	}

	if resp.Error == cues.CueListNotFound {
		return nil, ErrCueListNotFound
	}

	return nil, fmt.Errorf("failed to enumerate cues: %s", resp.Error)
}

func (c *Client) GetCueByNumber(cueListId string, cueNumber float64) (*types.Cue, error) {
	resp, err := messaging.Request[cues.GetCueByNumberResponse](c.messenger, cues.GetCueByNumberRequestSubject, &cues.GetCueByNumberRequest{
		CueListId: cueListId,
		CueNumber: cueNumber,
	})
	if err != nil {
		return nil, err
	}
	if resp.Success {
		return &resp.Response.Cue, nil
	}

	if resp.Error == cues.CueNotFound {
		return nil, ErrCueNotFound
	}

	if resp.Error == cues.CueListNotFound {
		return nil, ErrCueListNotFound
	}

	return nil, fmt.Errorf("failed to get cue: %s", resp.Error)
}

func (c *Client) GetCueById(cueListId string, cueId string) (*types.Cue, error) {
	resp, err := messaging.Request[cues.GetCueByIdResponse](c.messenger, cues.GetCueByIdRequestSubject, &cues.GetCueByIdRequest{
		CueId: cueId,
	})
	if err != nil {
		return nil, err
	}
	if resp.Success {
		return &resp.Response.Cue, nil
	}

	if resp.Error == cues.CueNotFound {
		return nil, ErrCueNotFound
	}

	if resp.Error == cues.CueListNotFound {
		return nil, ErrCueListNotFound
	}

	return nil, fmt.Errorf("failed to get cue: %s", resp.Error)
}

func (c *Client) UpdateCueAttributes(cueListId string, cueId string, field string, value any) error {
	resp, err := messaging.Request[cues.UpdateCueAttributesResponse](c.messenger, cues.UpdateCueAttributesRequestSubject, &cues.UpdateCueAttributesRequest{
		CueId: cueId,
		Field: field,
		Value: value,
	})
	if err != nil {
		return err
	}

	if resp.Success {
		return nil
	}

	if resp.Error == cues.CueNotFound {
		return ErrCueNotFound
	}

	if resp.Error == cues.CueListNotFound {
		return ErrCueListNotFound
	}

	return fmt.Errorf("failed to update cue attributes: %s", resp.Error)
}

func (c *Client) DeleteCue(cueId string) error {
	resp, err := messaging.Request[cues.DeleteCueResponse](c.messenger, cues.DeleteCueRequestSubject, &cues.DeleteCueRequest{
		CueId: cueId,
	})
	if err != nil {
		return fmt.Errorf("failed to delete cue: %w", err)
	}

	if resp.Success {
		return nil
	}

	if resp.Error == cues.CueNotFound {
		return ErrCueNotFound
	}

	return fmt.Errorf("failed to delete cue: %s", resp.Error)
}

func (c *Client) OnCueCreated(handler EventCallback[cues.CueCreatedEvent]) error {
	err := messaging.Subscribe[cues.CueCreatedEvent](c.messenger, false, cues.CueCreatedEventSubject, func(s string, e *cues.CueCreatedEvent) {
		handler(s, e)
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe to cue creation events: %w", err)
	}
	return nil
}

func (c *Client) OnCueAttributesUpdated(handler EventCallback[cues.CueUpdatedEvent]) error {
	err := messaging.Subscribe[cues.CueUpdatedEvent](c.messenger, false, cues.CueAttributesUpdatedEventSubject, func(s string, e *cues.CueUpdatedEvent) {
		handler(s, e)
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe to cue attributes update events: %w", err)
	}
	return nil
}

func (c *Client) OnCueDeleted(handler EventCallback[cues.CueDeletedEvent]) error {
	err := messaging.Subscribe[cues.CueDeletedEvent](c.messenger, false, cues.DeleteCueEventSubject, func(s string, e *cues.CueDeletedEvent) {
		handler(s, e)
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe to cue deletion events: %w", err)
	}
	return nil
}
