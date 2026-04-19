// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package client

import (
	"errors"
	"fmt"

	"github.com/stexxo/dynocue/components/system"
	"github.com/stexxo/dynocue/core/messaging"
)

var NoSaveLocation = errors.New("no save location provided")

func (c *Client) SaveShow(location string) error {
	resp, err := messaging.Request[system.PersistenceSaveResponse](c.messenger, system.PersistenceSaveRequestSubject, system.PersistenceSaveRequest{Location: location})
	if err != nil {
		return err
	}

	if resp.Success {
		return nil
	}

	if resp.Error == system.NoSaveLocation {
		return NoSaveLocation
	}

	return errors.New("save operation failed")
}

func (c *Client) OpenShow(location string) error {
	resp, err := messaging.Request[system.PersistenceOpenResponse](c.messenger, system.PersistenceOpenShowRequestSubject, system.PersistenceOpenRequest{Location: location})
	if err != nil {
		return err
	}

	if resp.Success {
		return nil
	}

	return fmt.Errorf("failed to open show from save location, %s", location)
}

func (c *Client) NewShow() error {
	resp, err := messaging.Request[system.PersistenceNewResponse](c.messenger, system.PersistenceNewShowRequestSubject, system.PersistenceNewRequest{})
	if err != nil {
		return err
	}
	if resp.Success {
		return nil
	}

	return fmt.Errorf("failed to create new show")
}

func (c *Client) HandleSaveEvent(fn func()) error {
	return messaging.Subscribe(c.messenger, false, system.PersistenceShowSavedEventSubject, func(s string, t *string) { fn() })
}

func (c *Client) HandleShowLoaded(fn func()) error {
	return messaging.Subscribe(c.messenger, false, system.PersistenceShowLoadedEventSubject, func(s string, t *string) { fn() })
}
