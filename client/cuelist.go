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

func (c *Client) CreateCueList(num uint, cueListType string) (uint, error) {
	resp, err := messaging.Request[cues.CreateCueListResponse](c.messenger, cues.CreateCueListRequestSubject, &cues.CreateCueListRequest{Number: num, CueListType: cueListType})
	if err != nil {
		return 0, err
	}
	if resp.Success {
		return resp.Response.Number, nil
	}

	if resp.Error == cues.CueListNumberExists {
		return 0, ErrCueListExists
	}

	return 0, fmt.Errorf("failed to create cue list: %s", resp.Error)
}

func (c *Client) EnumerateCueLists() ([]types.CueList, error) {
	resp, err := messaging.Request[cues.EnumerateCueListsResponse](c.messenger, cues.EnumerateCueListsRequestSubject, &cues.EnumerateCueListsRequest{})
	if err != nil {
		return nil, err
	}

	if resp.Success {
		return resp.Response.CueLists, nil
	}

	return nil, fmt.Errorf("failed to enumerate cue lists: %s", resp.Error)
}

func (c *Client) GetCueListByNumber(number uint) (*types.CueList, error) {
	resp, err := messaging.Request[cues.GetCueListByNumberResponse](c.messenger, cues.GetCueListByNumberRequestSubject, &cues.GetCueListByNumberRequest{Number: float64(number)})
	if err != nil {
		return nil, err
	}
	if resp.Success {
		return &resp.Response.CueList, nil
	}

	if resp.Error == cues.CueListNotFound {
		return nil, ErrCueListNotFound
	}

	return nil, fmt.Errorf("failed to get cue list: %s", resp.Error)
}

func (c *Client) GetCueListById(id string) (*types.CueList, error) {
	resp, err := messaging.Request[cues.GetCueListByIdResponse](c.messenger, cues.GetCueListByIdRequestSubject, &cues.GetCueListByIdRequest{Id: id})
	if err != nil {
		return nil, err
	}
	if resp.Success {
		return &resp.Response.CueList, nil
	}

	if resp.Error == cues.CueListNotFound {
		return nil, ErrCueListNotFound
	}

	return nil, fmt.Errorf("failed to get cue list: %s", resp.Error)
}

func (c *Client) UpdateCueListField(id string, field string, value any) error {
	resp, err := messaging.Request[cues.UpdateCueListAttributesResponse](c.messenger, cues.UpdateCueListAttributesRequestSubject, &cues.UpdateCueListAttributesRequest{Id: id, Field: field, Value: value})
	if err != nil {
		return err
	}

	if resp.Success {
		return nil
	}

	if resp.Error == cues.CueListNotFound {
		return ErrCueListNotFound
	}

	return fmt.Errorf("failed to update cue list attributes: %s", resp.Error)
}

func (c *Client) DeleteCueList(id string) error {
	resp, err := messaging.Request[cues.DeleteCueListsResponse](c.messenger, cues.DeleteCueListRequestSubject, &cues.DeleteCueListsRequest{Id: id})
	if err != nil {
		return fmt.Errorf("failed to delete cue list: %w", err)
	}

	if resp.Success {
		return nil
	}

	if resp.Error == cues.CueListNotFound {
		return ErrCueListNotFound
	}

	return fmt.Errorf("failed to delete cue list: %s", resp.Error)
}

func (c *Client) OnCueListCreated(handler EventCallback[cues.CueListCreatedEvent]) error {
	err := messaging.Subscribe[cues.CueListCreatedEvent](c.messenger, false, cues.CueListCreatedEventSubject, func(s string, c *cues.CueListCreatedEvent) {
		handler(s, c)
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe to cue list creation events: %w", err)
	}
	return nil
}

func (c *Client) OnCueListAttributesUpdated(handler EventCallback[cues.CueListAttributesUpdatedEvent]) error {
	err := messaging.Subscribe[cues.CueListAttributesUpdatedEvent](c.messenger, false, cues.CueListAttributesUpdatedEventSubject, func(s string, c *cues.CueListAttributesUpdatedEvent) {
		handler(s, c)
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe to cue list attributes update events: %w", err)
	}
	return nil
}

func (c *Client) OnCueListDeleted(handler EventCallback[cues.CueListDeletedEvent]) error {
	err := messaging.Subscribe[cues.CueListDeletedEvent](c.messenger, false, cues.DeleteCueListEventSubject, func(s string, c *cues.CueListDeletedEvent) {
		handler(s, c)
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe to cue list deletion events: %w", err)
	}
	return nil
}
