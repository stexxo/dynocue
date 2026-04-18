package client

import (
	"errors"

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
