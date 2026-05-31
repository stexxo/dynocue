// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package client

import (
	"fmt"

	"github.com/stexxo/dynocue/components/audio/api"
	"github.com/stexxo/dynocue/core/messaging"
)

func (c *Client) GetAudioOutputDevices() ([]api.AudioOutputDevice, error) {
	resp, err := messaging.Request[api.GetAudioOutputDevicesResponse](c.messenger, api.GetOutputDevicesRequestSubject, &api.GetAudioOutputDevicesRequest{})
	if err != nil {
		return nil, err
	}

	if resp.Success {
		return resp.Response.Devices, nil
	}

	return nil, fmt.Errorf("failed to get audio output devices: %s", resp.Error)
}
