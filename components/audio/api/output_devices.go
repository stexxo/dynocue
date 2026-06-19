package api

import (
	"errors"
	"fmt"

	"github.com/stexxo/dynocue/components/audio/model"
	"github.com/stexxo/dynocue/components/audio/types"
	"github.com/stexxo/dynocue/core/messaging"
	"github.com/stexxo/dynocue/pkg/vlc"
)

func (a *AudioAPI) registerOutputDevicesApis() error {
	return errors.Join(
		messaging.Reply[GetAudioOutputDevicesRequest, GetAudioOutputDevicesResponse](a.messenger, true, GetOutputDevicesRequestSubject, a.GetAudioDevices),
		messaging.Reply[CreateAudioOutputRequest, CreateAudioOutputResponse](a.messenger, true, CreateAudioOutputRequestSubject, a.CreateAudioOutput),
		messaging.Reply[GetAudioOutputRequest, GetAudioOutputResponse](a.messenger, true, GetAudioOutputRequestSubject, a.GetAudioOutput),
		messaging.Reply[EnumerateAudioOutputsRequest, EnumerateAudioOutputsResponse](a.messenger, true, EnumerateAudioOutputsRequestSubject, a.EnumerateAudioOutputs),
		messaging.Reply[DeleteAudioOutputRequest, DeleteAudioOutputResponse](a.messenger, true, DeleteAudioOutputRequestSubject, a.DeleteAudioOutput),
	)
}

const (
	GetOutputDevicesRequestSubject      = "request.audio.devices.get"
	CreateAudioOutputRequestSubject     = "request.audio.outputs.create"
	GetAudioOutputRequestSubject        = "request.audio.outputs.get"
	EnumerateAudioOutputsRequestSubject = "request.audio.outputs.enumerate"
	DeleteAudioOutputRequestSubject     = "request.audio.outputs.delete"

	AudioOutputNotFound     = "Audio output not found"
	AudioOutputNumberExists = "Audio output number already exists"
)

type GetAudioOutputDevicesRequest struct{}
type GetAudioOutputDevicesResponse struct {
	Devices []types.AudioDevice `json:"devices" msgpack:"devices"`
}

func (a *AudioAPI) GetAudioDevices(sub string, req *GetAudioOutputDevicesRequest) (*GetAudioOutputDevicesResponse, error) {
	devices, err := vlc.GetDevices()
	if err != nil {
		return nil, err
	}

	out := make([]types.AudioDevice, 0, len(devices))
	for _, device := range devices {
		out = append(out, types.AudioDevice{
			Name:         device.Name,
			FriendlyName: device.FriendlyName,
			Description:  device.Description,
		})
	}

	return &GetAudioOutputDevicesResponse{Devices: out}, nil
}

type CreateAudioOutputRequest struct {
	Number           uint   `json:"number" msgpack:"number" validate:"gte=0"`
	Label            string `json:"label" msgpack:"label"`
	AudioDeviceName  string `json:"audioDeviceName" msgpack:"audioDeviceName"`
	UseSystemDefault bool   `json:"useSystemDefault" msgpack:"useSystemDefault"`
}

type CreateAudioOutputResponse struct {
	Number uint `json:"number" msgpack:"number"`
}

func (a *AudioAPI) CreateAudioOutput(sub string, req *CreateAudioOutputRequest) (*CreateAudioOutputResponse, error) {
	var (
		number uint
		err    error
	)
	if req.UseSystemDefault {
		number, err = a.model.CreateAudioOutputWithSystemDefault(req.Number, req.Label)
	} else {
		number, err = a.model.CreateAudioOutputWithDeviceName(req.Number, req.Label, req.AudioDeviceName)
	}
	if err != nil {
		if errors.Is(err, model.ErrNumberExists) {
			return nil, errors.Join(err, &messaging.FriendlyError{FriendlyErr: AudioOutputNumberExists})
		}
		return nil, fmt.Errorf("failed to create audio output: %w", err)
	}
	return &CreateAudioOutputResponse{Number: number}, nil
}

type GetAudioOutputRequest struct {
	OutputId string `json:"outputId" msgpack:"outputId" validate:"required"`
}

type GetAudioOutputResponse struct {
	Output types.AudioOutput `json:"output" msgpack:"output"`
}

func (a *AudioAPI) GetAudioOutput(sub string, req *GetAudioOutputRequest) (*GetAudioOutputResponse, error) {
	out, err := a.model.GetAudioOutput(req.OutputId)
	if err != nil {
		return nil, fmt.Errorf("failed to get audio output: %w", err)
	}
	if out == nil {
		return nil, &messaging.FriendlyError{FriendlyErr: AudioOutputNotFound}
	}
	return &GetAudioOutputResponse{Output: *out}, nil
}

type EnumerateAudioOutputsRequest struct{}

type EnumerateAudioOutputsResponse struct {
	Outputs []types.AudioOutput `json:"outputs" msgpack:"outputs"`
}

func (a *AudioAPI) EnumerateAudioOutputs(sub string, req *EnumerateAudioOutputsRequest) (*EnumerateAudioOutputsResponse, error) {
	outputs, err := a.model.EnumerateAudioOutputs()
	if err != nil {
		return nil, fmt.Errorf("failed to enumerate audio outputs: %w", err)
	}
	return &EnumerateAudioOutputsResponse{Outputs: outputs}, nil
}

type DeleteAudioOutputRequest struct {
	OutputId string `json:"outputId" msgpack:"outputId" validate:"required"`
}

type DeleteAudioOutputResponse struct{}

func (a *AudioAPI) DeleteAudioOutput(sub string, req *DeleteAudioOutputRequest) (*DeleteAudioOutputResponse, error) {
	if err := a.model.DeleteAudioOutput(req.OutputId); err != nil {
		return nil, fmt.Errorf("failed to delete audio output: %w", err)
	}
	return &DeleteAudioOutputResponse{}, nil
}
