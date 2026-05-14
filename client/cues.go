// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package client

import (
	"errors"
	"fmt"

	"github.com/stexxo/dynocue/components/cues/api"
	"github.com/stexxo/dynocue/components/cues/types"
	"github.com/stexxo/dynocue/core/messaging"
)

var ErrCueExists = errors.New("cue with provided number already exists")
var ErrCueNotFound = errors.New("cue not found")

func (c *Client) CreateCue(cueListId string, cueNumber uint) (uint, error) {
	resp, err := messaging.Request[api.CreateCueResponse](c.messenger, api.CreateCueRequestSubject, &api.CreateCueRequest{
		CueListId: cueListId,
		Number:    cueNumber,
	})
	if err != nil {
		return 0, err
	}
	if resp.Success {
		return resp.Response.Number, nil
	}

	if resp.Error == api.CueNumberExists {
		return 0, ErrCueExists
	}

	return 0, fmt.Errorf("failed to create cue: %s", resp.Error)
}

func (c *Client) EnumerateCues(cueListId string) ([]types.Cue, error) {
	resp, err := messaging.Request[api.EnumerateCuesResponse](c.messenger, api.EnumerateCuesRequestSubject, &api.EnumerateCuesRequest{
		CueListId: cueListId,
	})
	if err != nil {
		return nil, err
	}

	if resp.Success {
		return resp.Response.Cues, nil
	}

	if resp.Error == api.CueListNotFound {
		return nil, ErrCueListNotFound
	}

	return nil, fmt.Errorf("failed to enumerate cues: %s", resp.Error)
}

func (c *Client) GetCueByNumber(cueListId string, cueNumber float64) (*types.Cue, error) {
	resp, err := messaging.Request[api.GetCueByNumberResponse](c.messenger, api.GetCueByNumberRequestSubject, &api.GetCueByNumberRequest{
		CueListId: cueListId,
		Number:    cueNumber,
	})
	if err != nil {
		return nil, err
	}
	if resp.Success {
		return &resp.Response.Cue, nil
	}

	if resp.Error == api.CueNotFound {
		return nil, ErrCueNotFound
	}

	if resp.Error == api.CueListNotFound {
		return nil, ErrCueListNotFound
	}

	return nil, fmt.Errorf("failed to get cue: %s", resp.Error)
}

func (c *Client) GetCueById(cueListId string, cueId string) (*types.Cue, error) {
	resp, err := messaging.Request[api.GetCueByIdResponse](c.messenger, api.GetCueByIdRequestSubject, &api.GetCueByIdRequest{
		CueId: cueId,
	})
	if err != nil {
		return nil, err
	}
	if resp.Success {
		return &resp.Response.Cue, nil
	}

	if resp.Error == api.CueNotFound {
		return nil, ErrCueNotFound
	}

	if resp.Error == api.CueListNotFound {
		return nil, ErrCueListNotFound
	}

	return nil, fmt.Errorf("failed to get cue: %s", resp.Error)
}

func (c *Client) UpdateCueAttributes(cueListId string, cueId string, field string, value any) error {
	resp, err := messaging.Request[api.UpdateCueAttributesResponse](c.messenger, api.UpdateCueAttributesRequestSubject, &api.UpdateCueAttributesRequest{
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

	if resp.Error == api.CueNotFound {
		return ErrCueNotFound
	}

	if resp.Error == api.CueListNotFound {
		return ErrCueListNotFound
	}

	return fmt.Errorf("failed to update cue attributes: %s", resp.Error)
}

func (c *Client) DeleteCue(cueId string) error {
	resp, err := messaging.Request[api.DeleteCueResponse](c.messenger, api.DeleteCueRequestSubject, &api.DeleteCueRequest{
		CueId: cueId,
	})
	if err != nil {
		return fmt.Errorf("failed to delete cue: %w", err)
	}

	if resp.Success {
		return nil
	}

	if resp.Error == api.CueNotFound {
		return ErrCueNotFound
	}

	return fmt.Errorf("failed to delete cue: %s", resp.Error)
}

func (c *Client) OnCueCreated(handler EventCallback[api.CueChangeEvent]) error {
	err := messaging.Subscribe[api.CueChangeEvent](c.messenger, false, api.CueCreatedEventSubject, func(s string, e *api.CueChangeEvent) {
		handler(s, e)
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe to cue creation events: %w", err)
	}
	return nil
}

func (c *Client) OnCueAttributesUpdated(handler EventCallback[api.CueChangeEvent]) error {
	err := messaging.Subscribe[api.CueChangeEvent](c.messenger, false, api.CueAttributesUpdatedEventSubject, func(s string, e *api.CueChangeEvent) {
		handler(s, e)
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe to cue attributes update events: %w", err)
	}
	return nil
}

func (c *Client) OnCueDeleted(handler EventCallback[api.CueChangeEvent]) error {
	err := messaging.Subscribe[api.CueChangeEvent](c.messenger, false, api.DeleteCueEventSubject, func(s string, e *api.CueChangeEvent) {
		handler(s, e)
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe to cue deletion events: %w", err)
	}
	return nil
}
