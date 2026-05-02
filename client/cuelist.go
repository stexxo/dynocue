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

func (c *Client) EnumerateCueLists() ([]types.CueListAttributes, error) {
	resp, err := messaging.Request[cues.EnumerateCueListsResponse](c.messenger, cues.EnumerateCueListsRequestSubject, &cues.EnumerateCueListsRequest{})
	if err != nil {
		return nil, err
	}

	if resp.Success {
		return resp.Response.CueLists, nil
	}

	return nil, fmt.Errorf("failed to enumerate cue lists: %s", resp.Error)
}

func (c *Client) GetCueListByNumber(number float64) (*types.CueListAttributes, error) {
	resp, err := messaging.Request[cues.GetCueListByNumberResponse](c.messenger, cues.GetCueListByNumberRequestSubject, &cues.GetCueListByNumberRequest{Number: number})
	if err != nil {
		return nil, err
	}
	if resp.Success {
		return &resp.Response.Attributes, nil
	}

	if resp.Error == cues.CueListNotFound {
		return nil, ErrCueListNotFound
	}

	return nil, fmt.Errorf("failed to get cue list: %s", resp.Error)
}

func (c *Client) GetCueListById(id string) (*types.CueListAttributes, error) {
	resp, err := messaging.Request[cues.GetCueListByIdResponse](c.messenger, cues.GetCueListByIdRequestSubject, &cues.GetCueListByIdRequest{Id: id})
	if err != nil {
		return nil, err
	}
	if resp.Success {
		return &resp.Response.Attributes, nil
	}

	if resp.Error == cues.CueListNotFound {
		return nil, ErrCueListNotFound
	}

	return nil, fmt.Errorf("failed to get cue list: %s", resp.Error)
}

func (c *Client) UpdateCueListField(id string, field string, value interface{}) (*types.CueListAttributes, error) {
	resp, err := messaging.Request[cues.UpdateCueListAttributesResponse](c.messenger, cues.UpdateCueListAttributesRequestSubject, &cues.UpdateCueListAttributesRequest{Id: id, Field: field, Value: value})
	if err != nil {
		return nil, err
	}

	if resp.Success {
		return &resp.Response.Attributes, nil
	}

	if resp.Error == cues.CueListNotFound {
		return nil, ErrCueListNotFound
	}

	return nil, fmt.Errorf("failed to update cue list attributes: %s", resp.Error)
}

func (c *Client) RenumberCueList(id string, newNumber float64) error {
	resp, err := messaging.Request[cues.RenumberCueListsResponse](c.messenger, cues.RenumberCueListRequestSubject, &cues.RenumberCueListsRequest{Id: id, NewNumber: newNumber})
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

func (c *Client) DeleteCueList(id string) error {
	err := messaging.Publish(c.messenger, cues.DeleteCueListRequestSubject, &cues.DeleteCueListsRequest{Id: id})
	if err != nil {
		return fmt.Errorf("failed to publish delete cue list request: %w", err)
	}
	return nil
}

func (c *Client) OnCueListCreated(handler EventCallback[types.CueListAttributes]) error {
	err := messaging.Subscribe[cues.CueListCreatedEvent](c.messenger, false, cues.CueListCreatedEventSubject, func(s string, c *cues.CueListCreatedEvent) {
		handler(s, &c.Attributes)
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe to cue list creation events: %w", err)
	}
	return nil
}

func (c *Client) OnCueListAttributesUpdated(handler EventCallback[types.CueListAttributes]) error {
	err := messaging.Subscribe[cues.CueListAttributesUpdatedEvent](c.messenger, false, cues.CueListAttributesUpdatedEventSubject, func(s string, c *cues.CueListAttributesUpdatedEvent) {
		handler(s, &c.Attributes)
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe to cue list attributes update events: %w", err)
	}
	return nil
}

type RenumberEvent struct {
	Id        string  `json:"id"`
	NewNumber float64 `json:"newNumber"`
}

func (c *Client) OnCueListRenumber(handler EventCallback[RenumberEvent]) error {
	err := messaging.Subscribe[cues.RenumberCueListEvent](c.messenger, false, cues.RenumberCueListEventSubject, func(s string, c *cues.RenumberCueListEvent) {
		handler(s, &RenumberEvent{
			Id:        c.Id,
			NewNumber: c.NewNumber,
		})
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe to cue list renumber events: %w", err)
	}
	return nil
}

func (c *Client) OnCueListDeleted(handler EventCallback[string]) error {
	err := messaging.Subscribe[cues.CueListDeletedEvent](c.messenger, false, cues.DeleteCueListEventSubject, func(s string, c *cues.CueListDeletedEvent) {
		handler(s, &c.Id)
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe to cue list deletion events: %w", err)
	}
	return nil
}
