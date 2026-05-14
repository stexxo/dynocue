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

var ErrCueListExists = errors.New("cue list with provided number already exists")
var ErrCueListNotFound = errors.New("cue list not found")

func (c *Client) CreateCueList(num uint, cueListType string) (uint, error) {
	resp, err := messaging.Request[api.CreateCueListResponse](c.messenger, api.CreateCueListRequestSubject, &api.CreateCueListRequest{Number: num, CueListType: cueListType})
	if err != nil {
		return 0, err
	}
	if resp.Success {
		return resp.Response.Number, nil
	}

	if resp.Error == api.CueListNumberExists {
		return 0, ErrCueListExists
	}

	return 0, fmt.Errorf("failed to create cue list: %s", resp.Error)
}

func (c *Client) EnumerateCueLists() ([]types.CueList, error) {
	resp, err := messaging.Request[api.EnumerateCueListsResponse](c.messenger, api.EnumerateCueListsRequestSubject, &api.EnumerateCueListsRequest{})
	if err != nil {
		return nil, err
	}

	if resp.Success {
		return resp.Response.CueLists, nil
	}

	return nil, fmt.Errorf("failed to enumerate cue lists: %s", resp.Error)
}

func (c *Client) GetCueListByNumber(number uint) (*types.CueList, error) {
	resp, err := messaging.Request[api.GetCueListByNumberResponse](c.messenger, api.GetCueListByNumberRequestSubject, &api.GetCueListByNumberRequest{Number: float64(number)})
	if err != nil {
		return nil, err
	}
	if resp.Success {
		return &resp.Response.CueList, nil
	}

	if resp.Error == api.CueListNotFound {
		return nil, ErrCueListNotFound
	}

	return nil, fmt.Errorf("failed to get cue list: %s", resp.Error)
}

func (c *Client) GetCueListById(id string) (*types.CueList, error) {
	resp, err := messaging.Request[api.GetCueListByIdResponse](c.messenger, api.GetCueListByIdRequestSubject, &api.GetCueListByIdRequest{CueListId: id})
	if err != nil {
		return nil, err
	}
	if resp.Success {
		return &resp.Response.CueList, nil
	}

	if resp.Error == api.CueListNotFound {
		return nil, ErrCueListNotFound
	}

	return nil, fmt.Errorf("failed to get cue list: %s", resp.Error)
}

func (c *Client) UpdateCueListField(id string, field string, value any) error {
	resp, err := messaging.Request[api.UpdateCueListAttributesResponse](c.messenger, api.UpdateCueListAttributesRequestSubject, &api.UpdateCueListAttributesRequest{CueListId: id, Field: field, Value: value})
	if err != nil {
		return err
	}

	if resp.Success {
		return nil
	}

	if resp.Error == api.CueListNotFound {
		return ErrCueListNotFound
	}

	return fmt.Errorf("failed to update cue list attributes: %s", resp.Error)
}

func (c *Client) DeleteCueList(id string) error {
	resp, err := messaging.Request[api.DeleteCueListResponse](c.messenger, api.DeleteCueListRequestSubject, &api.DeleteCueListRequest{CueListId: id})
	if err != nil {
		return fmt.Errorf("failed to delete cue list: %w", err)
	}

	if resp.Success {
		return nil
	}

	if resp.Error == api.CueListNotFound {
		return ErrCueListNotFound
	}

	return fmt.Errorf("failed to delete cue list: %s", resp.Error)
}

func (c *Client) OnCueListCreated(handler EventCallback[api.CueListChangeEvent]) error {
	err := messaging.Subscribe[api.CueListChangeEvent](c.messenger, false, api.CueListCreatedEventSubject, func(s string, c *api.CueListChangeEvent) {
		handler(s, c)
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe to cue list creation events: %w", err)
	}
	return nil
}

func (c *Client) OnCueListAttributesUpdated(handler EventCallback[api.CueListChangeEvent]) error {
	err := messaging.Subscribe[api.CueListChangeEvent](c.messenger, false, api.CueListAttributesUpdatedEventSubject, func(s string, c *api.CueListChangeEvent) {
		handler(s, c)
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe to cue list attributes update events: %w", err)
	}
	return nil
}

func (c *Client) OnCueListDeleted(handler EventCallback[api.CueListChangeEvent]) error {
	err := messaging.Subscribe[api.CueListChangeEvent](c.messenger, false, api.DeleteCueListEventSubject, func(s string, c *api.CueListChangeEvent) {
		handler(s, c)
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe to cue list deletion events: %w", err)
	}
	return nil
}
