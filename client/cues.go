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

func (c *Client) CreateCue(cueListId string, cueNumber float64) (float64, error) {
	resp, err := messaging.Request[cues.CreateCueResponse](c.messenger, cues.CreateCueRequestSubject, &cues.CreateCueRequest{
		CueListId: cueListId,
		CueNumber: cueNumber,
	})
	if err != nil {
		return -1, err
	}
	if resp.Success {
		return resp.Response.CueNumber, nil
	}

	if resp.Error == cues.CueNumberExists {
		return -1, ErrCueExists
	}

	return -1, fmt.Errorf("failed to create cue: %s", resp.Error)
}

func (c *Client) EnumerateCues(cueListId string) ([]types.CueMetadata, error) {
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

func (c *Client) GetCueByNumber(cueListNumber float64, cueNumber float64) (*types.CueMetadata, error) {
	resp, err := messaging.Request[cues.GetCueByNumberResponse](c.messenger, cues.GetCueByNumberRequestSubject, &cues.GetCueByNumberRequest{
		CueListNumber: cueListNumber,
		CueNumber:     cueNumber,
	})
	if err != nil {
		return nil, err
	}
	if resp.Success {
		return &resp.Response.Metadata, nil
	}

	if resp.Error == cues.CueNotFound {
		return nil, ErrCueNotFound
	}

	if resp.Error == cues.CueListNotFound {
		return nil, ErrCueListNotFound
	}

	return nil, fmt.Errorf("failed to get cue: %s", resp.Error)
}

func (c *Client) GetCueById(cueListId string, cueId string) (*types.CueMetadata, error) {
	resp, err := messaging.Request[cues.GetCueByIdResponse](c.messenger, cues.GetCueByIdRequestSubject, &cues.GetCueByIdRequest{
		CueListId: cueListId,
		CueId:     cueId,
	})
	if err != nil {
		return nil, err
	}
	if resp.Success {
		return &resp.Response.Metadata, nil
	}

	if resp.Error == cues.CueNotFound {
		return nil, ErrCueNotFound
	}

	if resp.Error == cues.CueListNotFound {
		return nil, ErrCueListNotFound
	}

	return nil, fmt.Errorf("failed to get cue: %s", resp.Error)
}

func (c *Client) UpdateCueLabel(cueListId string, cueId string, label string) (*types.CueMetadata, error) {
	resp, err := messaging.Request[cues.UpdateCueLabelResponse](c.messenger, cues.UpdateCueLabelRequestSubject, &cues.UpdateCueLabelRequest{
		CueListId: cueListId,
		CueId:     cueId,
		Label:     label,
	})
	if err != nil {
		return nil, err
	}

	if resp.Success {
		return &resp.Response.Metadata, nil
	}

	if resp.Error == cues.CueNotFound {
		return nil, ErrCueNotFound
	}

	if resp.Error == cues.CueListNotFound {
		return nil, ErrCueListNotFound
	}

	return nil, fmt.Errorf("failed to update cue label: %s", resp.Error)
}

func (c *Client) RenumberCue(cueListId string, cueId string, newNumber float64) error {
	resp, err := messaging.Request[cues.RenumberCueResponse](c.messenger, cues.RenumberCueRequestSubject, &cues.RenumberCueRequest{
		CueListId: cueListId,
		CueId:     cueId,
		NewNumber: newNumber,
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

	if resp.Error == cues.CueNumberExists {
		return ErrCueExists
	}

	return fmt.Errorf("failed to renumber cue: %s", resp.Error)
}

func (c *Client) DeleteCue(cueListId string, cueId string) error {
	err := messaging.Publish(c.messenger, cues.DeleteCueRequestSubject, &cues.DeleteCueRequest{
		CueListId: cueListId,
		CueId:     cueId,
	})
	if err != nil {
		return fmt.Errorf("failed to publish delete cue request: %w", err)
	}
	return nil
}

func (c *Client) OnCueCreated(handler EventCallback[types.CueMetadata]) error {
	err := messaging.Subscribe[cues.CueCreatedEvent](c.messenger, false, cues.CueCreatedEventSubject, func(s string, e *cues.CueCreatedEvent) {
		handler(s, &e.Metadata)
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe to cue creation events: %w", err)
	}
	return nil
}

func (c *Client) OnCueMetadataUpdated(handler EventCallback[types.CueMetadata]) error {
	err := messaging.Subscribe[cues.CueMetadataUpdatedEvent](c.messenger, false, cues.CueMetadataUpdatedEventSubject, func(s string, e *cues.CueMetadataUpdatedEvent) {
		handler(s, &e.Metadata)
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe to cue metadata update events: %w", err)
	}
	return nil
}

type CueRenumberEvent struct {
	CueListId string  `json:"cueListId"`
	CueId     string  `json:"cueId"`
	NewNumber float64 `json:"newNumber"`
}

func (c *Client) OnCueRenumber(handler EventCallback[CueRenumberEvent]) error {
	err := messaging.Subscribe[cues.RenumberCueEvent](c.messenger, false, cues.RenumberCueEventSubject, func(s string, e *cues.RenumberCueEvent) {
		handler(s, &CueRenumberEvent{
			CueListId: e.CueListId,
			CueId:     e.CueId,
			NewNumber: e.NewNumber,
		})
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe to cue renumber events: %w", err)
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
