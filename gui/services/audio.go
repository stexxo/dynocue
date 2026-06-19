// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package services

import (
	"errors"
	"path/filepath"

	"github.com/stexxo/dynocue/client"
	"github.com/stexxo/dynocue/components/audio/api"
	"github.com/stexxo/dynocue/components/audio/types"
	"github.com/stexxo/dynocue/core/logging"
	"github.com/wailsapp/wails/v3/pkg/application"
)

type AudioService struct {
	clientManager *client.Manager
	app           *application.App
	logger        logging.Logger
}

func NewAudioService(manager *client.Manager, app *application.App, logger logging.Logger) *AudioService {
	out := &AudioService{
		app:           app,
		logger:        logger,
		clientManager: manager,
	}
	out.clientManager.OnNewClient(out.onNewClient)

	app.Event.On("audio-file-drop", func(event *application.CustomEvent) {
		data, ok := event.Data.(map[string]string)
		if !ok {
			return
		}
		out.CreateAudioFile(data["file"])
	})

	return out
}

func (a *AudioService) onNewClient(cl *client.Client) error {
	return errors.Join(
		cl.OnAudioFileCreated(func(s string, t *api.FileChangeEvent) { a.app.Event.Emit(s, t) }),
		cl.OnAudioFileAttributesUpdated(func(s string, t *api.FileChangeEvent) { a.app.Event.Emit(s, t) }),
		cl.OnAudioFileDeleted(func(s string, e *api.FileChangeEvent) { a.app.Event.Emit(s, e) }),
	)
}

func (a *AudioService) CreateAudioFile(fileLocation string) (string, bool) {
	var out string
	err := a.clientManager.WithClient(func(c *client.Client) error {

		fileId, err := c.CreateAudioFile(fileLocation, filepath.Base(fileLocation), 0)
		if err != nil {
			return err
		}
		out = fileId
		return nil
	})

	if err != nil {
		a.logger.Error("failed to create audio file", "err", err)
		return "", false
	}

	return out, true
}

func (a *AudioService) ReplaceAudioFile(fileId string, fileLocation string) bool {
	err := a.clientManager.WithClient(func(c *client.Client) error {
		return c.ReplaceAudioFile(fileId, fileLocation)
	})

	if err != nil {
		a.logger.Error("failed to replace audio file", "err", err, "fileId", fileId)
		return false
	}

	return true
}

func (a *AudioService) EnumerateAudioFiles() ([]types.AudioFile, bool) {
	var out []types.AudioFile
	err := a.clientManager.WithClient(func(c *client.Client) error {
		files, err := c.EnumerateAudioFiles()
		if err != nil {
			return err
		}
		out = files
		return nil
	})

	if err != nil {
		a.logger.Error("failed to enumerate audio files", "err", err)
		return nil, false
	}

	return out, true
}

func (a *AudioService) GetAudioFile(fileId string) (*types.AudioFile, bool) {
	var out *types.AudioFile
	err := a.clientManager.WithClient(func(c *client.Client) error {
		file, err := c.GetAudioFile(fileId)
		if err != nil {
			return err
		}
		out = file
		return nil
	})

	if err != nil {
		a.logger.Error("failed to get audio file", "err", err, "fileId", fileId)
		return nil, false
	}

	return out, true
}

func (a *AudioService) UpdateAudioFileAttributes(fileId string, field string, value any) bool {
	err := a.clientManager.WithClient(func(c *client.Client) error {
		return c.UpdateAudioFileAttributes(fileId, field, value)
	})

	if err != nil {
		a.logger.Error("failed to update audio file attributes", "err", err, "fileId", fileId)
		return false
	}

	return true
}

func (a *AudioService) DeleteAudioFile(fileId string) bool {
	err := a.clientManager.WithClient(func(c *client.Client) error {
		return c.DeleteAudioFile(fileId)
	})

	if err != nil {
		a.logger.Error("failed to delete audio file", "err", err, "fileId", fileId)
		return false
	}

	return true
}

func (a *AudioService) AddAudioFile() bool {
	dia := a.app.Dialog.OpenFileWithOptions(&application.OpenFileDialogOptions{
		Title: "Add Audio File",
		Filters: []application.FileFilter{
			{
				DisplayName: "Audio Files",
				Pattern:     "*.wav;*.mp3;*.flac;*.ogg",
			},
		},
	})
	res, err := dia.PromptForSingleSelection()
	if err != nil {
		a.logger.Error("failed to open file dialog", "err", err)
		return false
	}
	if res == "" {
		return false
	}

	_, ok := a.CreateAudioFile(res)
	return ok
}

func (a *AudioService) ReplaceAudioFileWithDialog(fileId string) bool {
	dia := a.app.Dialog.OpenFileWithOptions(&application.OpenFileDialogOptions{
		Title: "Replace Audio File",
		Filters: []application.FileFilter{
			{
				DisplayName: "Audio Files",
				Pattern:     "*.wav;*.mp3;*.flac;*.ogg",
			},
		},
	})
	res, err := dia.PromptForSingleSelection()
	if err != nil {
		a.logger.Error("failed to open file dialog", "err", err)
		return false
	}
	if res == "" {
		return false
	}

	return a.ReplaceAudioFile(fileId, res)
}

func (a *AudioService) GetAudioOutputDevices() ([]types.AudioDevice, bool) {
	var out []types.AudioDevice
	err := a.clientManager.WithClient(func(c *client.Client) error {
		devices, err := c.GetAudioOutputDevices()
		if err != nil {
			return err
		}
		out = devices
		return nil
	})

	if err != nil {
		a.logger.Error("failed to get audio output devices", "err", err)
		return nil, false
	}

	return out, true
}

func (a *AudioService) CreateAudioOutputWithDeviceName(number uint, label string, audioDeviceName string) (uint, bool) {
	var out uint
	err := a.clientManager.WithClient(func(c *client.Client) error {
		n, err := c.CreateAudioOutputWithDeviceName(number, label, audioDeviceName)
		if err != nil {
			return err
		}
		out = n
		return nil
	})

	if err != nil {
		a.logger.Error("failed to create audio output", "err", err, "label", label, "deviceName", audioDeviceName)
		return 0, false
	}

	return out, true
}

func (a *AudioService) CreateAudioOutputWithSystemDefault(number uint, label string) (uint, bool) {
	var out uint
	err := a.clientManager.WithClient(func(c *client.Client) error {
		n, err := c.CreateAudioOutputWithSystemDefault(number, label)
		if err != nil {
			return err
		}
		out = n
		return nil
	})

	if err != nil {
		a.logger.Error("failed to create audio output with system default", "err", err, "label", label)
		return 0, false
	}

	return out, true
}

func (a *AudioService) GetAudioOutput(outputId string) (*types.AudioOutput, bool) {
	var out *types.AudioOutput
	err := a.clientManager.WithClient(func(c *client.Client) error {
		o, err := c.GetAudioOutput(outputId)
		if err != nil {
			return err
		}
		out = o
		return nil
	})

	if err != nil {
		a.logger.Error("failed to get audio output", "err", err, "outputId", outputId)
		return nil, false
	}

	return out, true
}

func (a *AudioService) EnumerateAudioOutputs() ([]types.AudioOutput, bool) {
	var out []types.AudioOutput
	err := a.clientManager.WithClient(func(c *client.Client) error {
		outputs, err := c.EnumerateAudioOutputs()
		if err != nil {
			return err
		}
		out = outputs
		return nil
	})

	if err != nil {
		a.logger.Error("failed to enumerate audio outputs", "err", err)
		return nil, false
	}

	return out, true
}

func (a *AudioService) DeleteAudioOutput(outputId string) bool {
	err := a.clientManager.WithClient(func(c *client.Client) error {
		return c.DeleteAudioOutput(outputId)
	})

	if err != nil {
		a.logger.Error("failed to delete audio output", "err", err, "outputId", outputId)
		return false
	}

	return true
}
