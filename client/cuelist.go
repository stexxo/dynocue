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

var ErrCueListExists = errors.New("cue list with provided number already exists")
var ErrCueListNotFound = errors.New("cue list not found")

func (c *Client) CreateCueList(num float64, cueListType string) (float64, error) {
	resp, err := messaging.Request[cues.CreateCueListResponse](c.messenger, cues.CreateCueListRequestSubject, &cues.CreateCueListRequest{Number: num, CueListType: cueListType})
	if err != nil {
		return -1, err
	}
	if resp.Success {
		return resp.Response.Number, nil
	}

	if resp.Error == cues.CueListNumberExists {
		return -1, ErrCueListExists
	}

	return -1, fmt.Errorf("failed to create cue list: %s", resp.Error)
}

func (c *Client) EnumerateCueLists() ([]types.CueListMetadata, error) {
	resp, err := messaging.Request[cues.EnumerateCueListsResponse](c.messenger, cues.EnumerateCueListsRequestSubject, &cues.EnumerateCueListsRequest{})
	if err != nil {
		return nil, err
	}

	if resp.Success {
		return resp.Response.CueLists, nil
	}

	return nil, fmt.Errorf("failed to enumerate cue lists: %s", resp.Error)
}

func (c *Client) GetCueList(number float64) (*types.CueListMetadata, error) {
	resp, err := messaging.Request[cues.GetCueListResponse](c.messenger, cues.GetCueListRequestSubject, &cues.GetCueListRequest{Number: number})
	if err != nil {
		return nil, err
	}
	if resp.Success {
		return &resp.Response.CueListMetadata, nil
	}

	if resp.Error == cues.CueListNotFound {
		return nil, ErrCueListNotFound
	}

	return nil, fmt.Errorf("failed to get cue list: %s", resp.Error)
}

func (c *Client) SetCueListLabel(num float64, label string) (*types.CueListMetadata, error) {
	resp, err := messaging.Request[cues.UpdateCueListLabelResponse](c.messenger, cues.UpdateCueListLabelRequestSubject, &cues.UpdateCueListLabelRequest{Number: num, Label: label})
	if err != nil {
		return nil, err
	}

	if resp.Success {
		return &resp.Response.Metadata, nil
	}

	if resp.Error == cues.CueListNotFound {
		return nil, ErrCueListNotFound
	}

	return nil, fmt.Errorf("failed to update cue list label: %s", resp.Error)
}

func (c *Client) RenumberCueList(originalNumber float64, newNumber float64) error {
	resp, err := messaging.Request[cues.RenumberCueListsResponse](c.messenger, cues.RenumberCueListRequestSubject, &cues.RenumberCueListsRequest{OriginalNumber: originalNumber, NewNumber: newNumber})
	if err != nil {
		return err
	}

	if resp.Success {
		return nil
	}

	if resp.Error == cues.CueListNotFound {
		return ErrCueListNotFound
	}

	if resp.Error == cues.CueListNumberExists {
		return ErrCueListExists
	}

	return fmt.Errorf("failed to renumber cue list: %s", resp.Error)
}

func (c *Client) DeleteCueList(number float64) error {
	err := messaging.Publish(c.messenger, cues.DeleteCueListRequestSubject, &cues.DeleteCueListsRequest{Number: number})
	if err != nil {
		return fmt.Errorf("failed to publish delete cue list request: %w", err)
	}
	return nil
}

func (c *Client) OnCueListCreated(handler EventCallback[types.CueListMetadata]) error {
	err := messaging.Subscribe[cues.CueListCreatedEvent](c.messenger, false, cues.CueListCreatedEventSubject, func(s string, c *cues.CueListCreatedEvent) {
		handler(s, &c.CueListMetadata)
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe to cue list creation events: %w", err)
	}
	return nil
}

func (c *Client) OnCueListMetadataUpdated(handler EventCallback[types.CueListMetadata]) error {
	err := messaging.Subscribe[cues.CueListMetadataUpdatedEvent](c.messenger, false, cues.CueListMetadataUpdatedEventSubject, func(s string, c *cues.CueListMetadataUpdatedEvent) {
		handler(s, &c.Metadata)
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe to cue list metadata update events: %w", err)
	}
	return nil
}

type RenumberEvent struct {
	OriginalNumber float64
	NewNumber      float64
}

func (c *Client) OnCueListRenumber(handler EventCallback[RenumberEvent]) error {
	err := messaging.Subscribe[cues.RenumberCueListEvent](c.messenger, false, cues.RenumberCueListEventSubject, func(s string, c *cues.RenumberCueListEvent) {
		handler(s, &RenumberEvent{
			OriginalNumber: c.OriginalNumber,
			NewNumber:      c.NewNumber,
		})
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe to cue list renumber events: %w", err)
	}
	return nil
}

func (c *Client) OnCueListDeleted(handler EventCallback[float64]) error {
	err := messaging.Subscribe[cues.CueListDeletedEvent](c.messenger, false, cues.DeleteCueListEventSubject, func(s string, c *cues.CueListDeletedEvent) {
		handler(s, &c.Number)
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe to cue list deletion events: %w", err)
	}
	return nil
}
