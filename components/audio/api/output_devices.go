package api

import (
	"errors"

	"github.com/stexxo/dynocue/core/messaging"
)

func (a *AudioAPI) registerOutputDevicesApis() error {
	return errors.Join(
		messaging.Reply[GetAudioOutputDevicesRequest, GetAudioOutputDevicesResponse](a.messenger, true, GetOutputDevicesRequestSubject, a.GetAudioOutputDevices))
}

const GetOutputDevicesRequestSubject = "audio.output_devices.get"

type GetAudioOutputDevicesRequest struct{}
type GetAudioOutputDevicesResponse struct {
	Devices []AudioOutputDevice `json:"devices" msgpack:"devices"`
}
type AudioOutputDevice struct {
	Id          string `json:"id" msgpack:"id"`
	Name        string `json:"name" msgpack:"name"`
	Description string `json:"description" msgpack:"description"`
}

func (a *AudioAPI) GetAudioOutputDevices(sub string, req *GetAudioOutputDevicesRequest) (*GetAudioOutputDevicesResponse, error) {
	return &GetAudioOutputDevicesResponse{Devices: []AudioOutputDevice{}}, nil
}
