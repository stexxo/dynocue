// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package client

import (
	"errors"
	"fmt"

	"github.com/stexxo/dynocue/components/audio/api"
	"github.com/stexxo/dynocue/components/audio/types"
	"github.com/stexxo/dynocue/core/messaging"
)

var (
	ErrAudioOutputNotFound     = errors.New("audio output not found")
	ErrAudioOutputNumberExists = errors.New("audio output with provided number already exists")
)

func (c *Client) GetAudioOutputDevices() ([]types.AudioDevice, error) {
	resp, err := messaging.Request[api.GetAudioOutputDevicesResponse](c.messenger, api.GetOutputDevicesRequestSubject, &api.GetAudioOutputDevicesRequest{})
	if err != nil {
		return nil, err
	}

	if resp.Success {
		return resp.Response.Devices, nil
	}

	return nil, fmt.Errorf("failed to get audio output devices: %s", resp.Error)
}

// CreateAudioOutputWithDeviceName creates a new audio output bound to a specific
// audio device by name. Pass number=0 to auto-assign the next available number.
func (c *Client) CreateAudioOutputWithDeviceName(number uint, label string, audioDeviceName string) (uint, error) {
	return c.createAudioOutput(&api.CreateAudioOutputRequest{
		Number:           number,
		Label:            label,
		AudioDeviceName:  audioDeviceName,
		UseSystemDefault: false,
	})
}

// CreateAudioOutputWithSystemDefault creates a new audio output that follows the
// system default audio device. Pass number=0 to auto-assign the next available number.
func (c *Client) CreateAudioOutputWithSystemDefault(number uint, label string) (uint, error) {
	return c.createAudioOutput(&api.CreateAudioOutputRequest{
		Number:           number,
		Label:            label,
		UseSystemDefault: true,
	})
}

func (c *Client) createAudioOutput(req *api.CreateAudioOutputRequest) (uint, error) {
	resp, err := messaging.Request[api.CreateAudioOutputResponse](c.messenger, api.CreateAudioOutputRequestSubject, req)
	if err != nil {
		return 0, err
	}
	if resp.Success {
		return resp.Response.Number, nil
	}

	if resp.Error == api.AudioOutputNumberExists {
		return 0, ErrAudioOutputNumberExists
	}

	return 0, fmt.Errorf("failed to create audio output: %s", resp.Error)
}

func (c *Client) GetAudioOutput(outputId string) (*types.AudioOutput, error) {
	resp, err := messaging.Request[api.GetAudioOutputResponse](c.messenger, api.GetAudioOutputRequestSubject, &api.GetAudioOutputRequest{
		OutputId: outputId,
	})
	if err != nil {
		return nil, err
	}
	if resp.Success {
		return &resp.Response.Output, nil
	}

	if resp.Error == api.AudioOutputNotFound {
		return nil, ErrAudioOutputNotFound
	}

	return nil, fmt.Errorf("failed to get audio output: %s", resp.Error)
}

func (c *Client) EnumerateAudioOutputs() ([]types.AudioOutput, error) {
	resp, err := messaging.Request[api.EnumerateAudioOutputsResponse](c.messenger, api.EnumerateAudioOutputsRequestSubject, &api.EnumerateAudioOutputsRequest{})
	if err != nil {
		return nil, err
	}
	if resp.Success {
		return resp.Response.Outputs, nil
	}

	return nil, fmt.Errorf("failed to enumerate audio outputs: %s", resp.Error)
}

func (c *Client) DeleteAudioOutput(outputId string) error {
	resp, err := messaging.Request[api.DeleteAudioOutputResponse](c.messenger, api.DeleteAudioOutputRequestSubject, &api.DeleteAudioOutputRequest{
		OutputId: outputId,
	})
	if err != nil {
		return err
	}
	if resp.Success {
		return nil
	}

	if resp.Error == api.AudioOutputNotFound {
		return ErrAudioOutputNotFound
	}

	return fmt.Errorf("failed to delete audio output: %s", resp.Error)
}
