// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package api

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/google/uuid"
	"github.com/stexxo/dynocue/components/audio/model"
	"github.com/stexxo/dynocue/components/audio/types"
	"github.com/stexxo/dynocue/core/messaging"
	"github.com/stexxo/dynocue/pkg/probe"
)

func (a *AudioAPI) registerFileApis() error {
	return errors.Join(
		messaging.Reply[UploadAudioFileRequest, UploadAudioFileResponse](a.messenger, true, CreateAudioFileRequestSubject, a.CreateAudioFile),
		messaging.Reply[ReplaceAudioFileRequest, ReplaceAudioFileResponse](a.messenger, true, ReplaceAudioFileRequestSubject, a.ReplaceAudioFile),
		messaging.Reply[EnumerateAudioFilesRequest, EnumerateAudioFilesResponse](a.messenger, true, EnumerateAudioFilesRequestSubject, a.EnumerateAudioFiles),
		messaging.Reply[GetAudioFileRequest, GetAudioFileResponse](a.messenger, true, GetAudioFileRequestSubject, a.GetAudioFile),
		messaging.Reply[UpdateAudioFileAttributesRequest, UpdateAudioFileAttributesResponse](a.messenger, true, UpdateAudioFileAttributesRequestSubject, a.UpdateAudioFileAttributes),
		messaging.Reply[DeleteAudioFileRequest, DeleteAudioFileResponse](a.messenger, true, DeleteAudioFileRequestSubject, a.DeleteAudioFile),
	)
}

const (
	CreateAudioFileRequestSubject = "request.audio.file.create"
	AudioFileNotFound             = "Audio file not found"
	AudioFileNumberExists         = "Audio file number already exists"
)

type UploadAudioFileRequest struct {
	LocationInTempBucket string `json:"locationInTempBucket" msgpack:"locationInTempBucket"`
	Label                string `json:"label" msgpack:"label"`
	Number               uint   `msgpack:"number" json:"number" validate:"gte=0"`
}

type UploadAudioFileResponse struct {
	FileId string `msgpack:"fileId" json:"fileId"`
	Number uint   `msgpack:"number" json:"number"`
}

func (a *AudioAPI) CreateAudioFile(sub string, req *UploadAudioFileRequest) (*UploadAudioFileResponse, error) {
	// Create a UUID for the file
	fileUUID := uuid.NewString()

	// Create a Key where the file will live in the Audio Object Store
	fileKey := fmt.Sprintf("files/%s", fileUUID)

	// Copy the file from the temporary location to the Audio Object Store
	err := a.persistence.CopyFromTempLocation(req.LocationInTempBucket, fileKey, true)
	if err != nil {
		return nil, fmt.Errorf("failed to copy file to object store: %w", err)
	}

	// Create a new AudioFile object in the database
	number, err := a.model.AddFile(fileUUID, fileKey, req.Label, req.Number)
	if err != nil {
		if errors.Is(err, model.ErrNumberExists) {
			return nil, errors.Join(err, &messaging.FriendlyError{FriendlyErr: AudioFileNumberExists})
		}
		return nil, fmt.Errorf("failed to add file to database: %w", err)
	}

	// Get a Temporary Location
	temp := path.Join(os.TempDir(), "dynocue", "audio")
	err = os.MkdirAll(temp, 0755)
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary directory: %w", err)
	}

	fullObjectKey := a.persistence.GetFullObjectKey(fileKey)
	err = a.persistence.ObjectStore().GetFile(context.Background(), fullObjectKey, path.Join(temp, fileUUID))
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve file from object store: %w", err)
	}
	defer os.Remove(path.Join(temp, fileUUID))

	info, err := probe.ProbeFile(path.Join(temp, fileUUID))
	if err != nil {
		return nil, fmt.Errorf("failed to probe file: %w", err)
	}

	err = a.model.SetFileProperties(fileUUID, info.Duration, uint(info.SizeBytes), info.Format)
	if err != nil {
		return nil, fmt.Errorf("failed to set file properties: %w", err)
	}

	return &UploadAudioFileResponse{FileId: fileUUID, Number: number}, nil
}

const ReplaceAudioFileRequestSubject = "request.audio.file.replace"

type ReplaceAudioFileRequest struct {
	FileId               string `json:"fileId" msgpack:"fileId" validate:"required"`
	LocationInTempBucket string `json:"locationInTempBucket" msgpack:"locationInTempBucket" validate:"required"`
}

type ReplaceAudioFileResponse struct{}

func (a *AudioAPI) ReplaceAudioFile(sub string, req *ReplaceAudioFileRequest) (*ReplaceAudioFileResponse, error) {
	f, err := a.model.GetFile(req.FileId)
	if err != nil {
		if errors.Is(err, model.ErrFileNotFound) {
			return nil, errors.Join(err, &messaging.FriendlyError{FriendlyErr: AudioFileNotFound})
		}
		return nil, fmt.Errorf("failed to get file from database: %w", err)
	}

	err = a.persistence.CopyFromTempLocation(req.LocationInTempBucket, f.Key, true)
	if err != nil {
		return nil, fmt.Errorf("failed to copy file from temp location: %w", err)
	}

	// Get a Temporary Location
	temp := path.Join(os.TempDir(), "dynocue", "audio")
	err = os.MkdirAll(temp, 0755)
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary directory: %w", err)
	}

	fullObjectKey := a.persistence.GetFullObjectKey(f.Key)
	err = a.persistence.ObjectStore().GetFile(context.Background(), fullObjectKey, path.Join(temp, f.Key))
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve file from object store: %w", err)
	}
	defer os.Remove(path.Join(temp, f.Key))

	info, err := probe.ProbeFile(path.Join(temp, f.Key))
	if err != nil {
		return nil, fmt.Errorf("failed to probe file: %w", err)
	}

	err = a.model.SetFileProperties(req.FileId, info.Duration, uint(info.SizeBytes), info.Format)
	if err != nil {
		return nil, fmt.Errorf("failed to set file properties: %w", err)
	}
	return &ReplaceAudioFileResponse{}, nil
}

const EnumerateAudioFilesRequestSubject = "request.audio.file.enumerate"

type EnumerateAudioFilesRequest struct{}

type EnumerateAudioFilesResponse struct {
	Files []types.AudioFile `msgpack:"files" json:"files"`
}

func (a *AudioAPI) EnumerateAudioFiles(sub string, req *EnumerateAudioFilesRequest) (*EnumerateAudioFilesResponse, error) {
	files, err := a.model.EnumerateFiles()
	if err != nil {
		return nil, fmt.Errorf("failed to list files from database: %w", err)
	}

	return &EnumerateAudioFilesResponse{Files: files}, nil
}

const GetAudioFileRequestSubject = "request.audio.file.get"

type GetAudioFileRequest struct {
	FileId string `json:"fileId" msgpack:"fileId" validate:"required"`
}

type GetAudioFileResponse struct {
	File types.AudioFile `msgpack:"file" json:"file"`
}

func (a *AudioAPI) GetAudioFile(sub string, req *GetAudioFileRequest) (*GetAudioFileResponse, error) {
	file, err := a.model.GetFile(req.FileId)
	if err != nil {
		if errors.Is(err, model.ErrFileNotFound) {
			return nil, errors.Join(err, &messaging.FriendlyError{FriendlyErr: AudioFileNotFound})
		}
		return nil, fmt.Errorf("failed to get file from database: %w", err)
	}

	return &GetAudioFileResponse{File: *file}, nil
}

const UpdateAudioFileAttributesRequestSubject = "request.audio.file.attributes.update"

type UpdateAudioFileAttributesRequest struct {
	FileId string      `json:"fileId" msgpack:"fileId" validate:"required"`
	Field  string      `json:"field" msgpack:"field" validate:"required,oneof=number label"`
	Value  interface{} `json:"value" msgpack:"value"`
}

type UpdateAudioFileAttributesResponse struct{}

func (a *AudioAPI) UpdateAudioFileAttributes(sub string, req *UpdateAudioFileAttributesRequest) (*UpdateAudioFileAttributesResponse, error) {
	err := a.model.UpdateFile(req.FileId, req.Field, req.Value)
	if err != nil {
		if errors.Is(err, model.ErrFileNotFound) {
			return nil, errors.Join(err, &messaging.FriendlyError{FriendlyErr: AudioFileNotFound})
		}
		return nil, fmt.Errorf("failed to update audio file attributes: %w", err)
	}

	return &UpdateAudioFileAttributesResponse{}, nil
}

const DeleteAudioFileRequestSubject = "request.audio.file.delete"

type DeleteAudioFileRequest struct {
	FileId string `json:"fileId" msgpack:"fileId" validate:"required"`
}

type DeleteAudioFileResponse struct {
}

func (a *AudioAPI) DeleteAudioFile(sub string, req *DeleteAudioFileRequest) (*DeleteAudioFileResponse, error) {
	err := a.model.DeleteFile(req.FileId)
	if err != nil {
		if errors.Is(err, model.ErrFileNotFound) {
			return nil, errors.Join(err, &messaging.FriendlyError{FriendlyErr: AudioFileNotFound})
		}
		return nil, fmt.Errorf("failed to delete file from database: %w", err)
	}

	return &DeleteAudioFileResponse{}, nil
}
