// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package client

import (
	"errors"
	"fmt"
	"os"

	"github.com/stexxo/dynocue/components/audio/api"
	"github.com/stexxo/dynocue/components/audio/types"
	"github.com/stexxo/dynocue/core/messaging"
)

var ErrAudioFileNotFound = errors.New("audio file not found")
var ErrAudioFileExists = errors.New("audio file with provided number already exists")

func (c *Client) CreateAudioFile(fileLocation string, number uint) (string, error) {
	file, err := os.Open(fileLocation)
	if err != nil {
		return "", fmt.Errorf("failed to open audio file: %w", err)
	}
	defer file.Close()

	key, err := c.persistence.WriteToTempLocation(file)
	if err != nil {
		return "", fmt.Errorf("failed to write audio file to temp location: %w", err)
	}

	resp, err := messaging.Request[api.UploadAudioFileResponse](c.messenger, api.CreateAudioFileRequestSubject, &api.UploadAudioFileRequest{
		LocationInTempBucket: key,
		Number:               number,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload audio file: %w", err)
	}
	if resp.Success {
		return resp.Response.FileId, nil
	}

	if resp.Error == api.AudioFileNumberExists {
		return "", ErrAudioFileExists
	}

	return "", fmt.Errorf("failed to create audio file: %s", resp.Error)
}

func (c *Client) ReplaceAudioFile(fileId string, fileLocation string) error {
	file, err := os.Open(fileLocation)
	if err != nil {
		return fmt.Errorf("failed to open audio file: %w", err)
	}
	defer file.Close()

	key, err := c.persistence.WriteToTempLocation(file)
	if err != nil {
		return fmt.Errorf("failed to write audio file to temp location: %w", err)
	}

	resp, err := messaging.Request[api.ReplaceAudioFileResponse](c.messenger, api.ReplaceAudioFileRequestSubject, &api.ReplaceAudioFileRequest{
		FileId:               fileId,
		LocationInTempBucket: key,
	})
	if err != nil {
		return err
	}
	if resp.Success {
		return nil
	}

	if resp.Error == api.AudioFileNotFound {
		return ErrAudioFileNotFound
	}

	return fmt.Errorf("failed to replace audio file: %s", resp.Error)
}

func (c *Client) EnumerateAudioFiles() ([]types.AudioFile, error) {
	resp, err := messaging.Request[api.EnumerateAudioFilesResponse](c.messenger, api.EnumerateAudioFilesRequestSubject, &api.EnumerateAudioFilesRequest{})
	if err != nil {
		return nil, err
	}
	if resp.Success {
		return resp.Response.Files, nil
	}

	return nil, fmt.Errorf("failed to enumerate audio files: %s", resp.Error)
}

func (c *Client) GetAudioFile(fileId string) (*types.AudioFile, error) {
	resp, err := messaging.Request[api.GetAudioFileResponse](c.messenger, api.GetAudioFileRequestSubject, &api.GetAudioFileRequest{
		FileId: fileId,
	})
	if err != nil {
		return nil, err
	}
	if resp.Success {
		return &resp.Response.File, nil
	}

	if resp.Error == api.AudioFileNotFound {
		return nil, ErrAudioFileNotFound
	}

	return nil, fmt.Errorf("failed to get audio file: %s", resp.Error)
}

func (c *Client) DeleteAudioFile(fileId string) error {
	resp, err := messaging.Request[api.DeleteAudioFileResponse](c.messenger, api.DeleteAudioFileRequestSubject, &api.DeleteAudioFileRequest{
		FileId: fileId,
	})
	if err != nil {
		return err
	}
	if resp.Success {
		return nil
	}

	if resp.Error == api.AudioFileNotFound {
		return ErrAudioFileNotFound
	}

	return fmt.Errorf("failed to delete audio file: %s", resp.Error)
}

func (c *Client) OnAudioFileCreated(handler EventCallback[api.FileChangeEvent]) error {
	err := messaging.Subscribe[api.FileChangeEvent](c.messenger, false, api.FileCreatedEventSubject, func(s string, e *api.FileChangeEvent) {
		handler(s, e)
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe to audio file creation events: %w", err)
	}
	return nil
}

func (c *Client) OnAudioFileAttributesUpdated(handler EventCallback[api.FileChangeEvent]) error {
	err := messaging.Subscribe[api.FileChangeEvent](c.messenger, false, api.FileAttributesUpdatedEventSubject, func(s string, e *api.FileChangeEvent) {
		handler(s, e)
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe to audio file attributes update events: %w", err)
	}
	return nil
}

func (c *Client) OnAudioFileDeleted(handler EventCallback[api.FileChangeEvent]) error {
	err := messaging.Subscribe[api.FileChangeEvent](c.messenger, false, api.DeleteFileEventSubject, func(s string, e *api.FileChangeEvent) {
		handler(s, e)
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe to audio file deletion events: %w", err)
	}
	return nil
}
