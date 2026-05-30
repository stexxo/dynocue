package api

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/stexxo/dynocue/components/audio/model"
	"github.com/stexxo/dynocue/components/audio/types"
	"github.com/stexxo/dynocue/core/messaging"
)

func (a *AudioAPI) registerFileApis() error {
	return errors.Join(
		messaging.Reply[UploadAudioFileRequest, UploadAudioFileResponse](a.messenger, true, CreateAudioFileRequestSubject, a.CreateAudioFile),
		messaging.Reply[ReplaceAudioFileRequest, ReplaceAudioFileResponse](a.messenger, true, ReplaceAudioFileRequestSubject, a.ReplaceAudioFile),
		messaging.Reply[EnumerateAudioFilesRequest, EnumerateAudioFilesResponse](a.messenger, true, EnumerateAudioFilesRequestSubject, a.EnumerateAudioFiles),
		messaging.Reply[GetAudioFileRequest, GetAudioFileResponse](a.messenger, true, GetAudioFileRequestSubject, a.GetAudioFile),
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
	number, err := a.model.AddFile(fileUUID, fileKey, req.Number)
	if err != nil {
		if errors.Is(err, model.ErrNumberExists) {
			return nil, errors.Join(err, &messaging.FriendlyError{FriendlyErr: AudioFileNumberExists})
		}
		return nil, fmt.Errorf("failed to add file to database: %w", err)
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
