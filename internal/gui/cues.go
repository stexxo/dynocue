package gui

import (
	"gitlab.com/stexxo/dynocue/api/cues"
)

func (c *Commands) CreateCue(input cues.CreateCueInput) (*cues.CreateCueOutput, error) {
	return makeRequest[cues.CreateCueInput, cues.CreateCueOutput](c, cues.RequestCreateCue, input)
}

func (c *Commands) UpdateCueMetadata(input cues.UpdateCueMetadataInput) (*cues.UpdateCueMetadataOutput, error) {
	return makeRequest[cues.UpdateCueMetadataInput, cues.UpdateCueMetadataOutput](c, cues.RequestUpdateCueMetadata, input)
}

func (c *Commands) GetCueMetadata(input cues.GetCueMetadataInput) (*cues.GetCueMetadataOutput, error) {
	return makeRequest[cues.GetCueMetadataInput, cues.GetCueMetadataOutput](c, cues.RequestGetCueMetadata, input)
}

func (c *Commands) EnumerateCue(input cues.EnumerateCueInput) (*cues.EnumerateCueOutput, error) {
	return makeRequest[cues.EnumerateCueInput, cues.EnumerateCueOutput](c, cues.RequestEnumerateCue, input)
}

func (c *Commands) DeleteCue(input cues.DeleteCueInput) (*cues.DeleteCueOutput, error) {
	return makeRequest[cues.DeleteCueInput, cues.DeleteCueOutput](c, cues.RequestDeleteCue, input)
}

func (c *Commands) MoveCue(input cues.MoveCueInput) (*cues.MoveCueOutput, error) {
	return makeRequest[cues.MoveCueInput, cues.MoveCueOutput](c, cues.RequestMoveCue, input)
}
