package gui

import (
	"gitlab.com/stexxo/dynocue/api/cues"
)

func (c *Commands) CreateCueList(input cues.CreateCueListInput) (*cues.CreateCueListOutput, error) {
	return makeRequest[cues.CreateCueListInput, cues.CreateCueListOutput](c, cues.RequestCreateCueList, input)
}

func (c *Commands) UpdateCueListMetadata(input cues.UpdateCueListMetadataInput) (*cues.UpdateCueListMetadataOutput, error) {
	return makeRequest[cues.UpdateCueListMetadataInput, cues.UpdateCueListMetadataOutput](c, cues.RequestUpdateCueListMetadata, input)
}

func (c *Commands) GetCueListMetadata(input cues.GetCueListMetadataInput) (*cues.GetCueListMetadataOutput, error) {
	return makeRequest[cues.GetCueListMetadataInput, cues.GetCueListMetadataOutput](c, cues.RequestGetCueListMetadata, input)
}

func (c *Commands) EnumerateCueList(input cues.EnumerateCueListInput) (*cues.EnumerateCueListOutput, error) {
	return makeRequest[cues.EnumerateCueListInput, cues.EnumerateCueListOutput](c, cues.RequestEnumerateCueList, input)
}

func (c *Commands) DeleteCueList(input cues.DeleteCueListInput) (*cues.DeleteCueListOutput, error) {
	return makeRequest[cues.DeleteCueListInput, cues.DeleteCueListOutput](c, cues.RequestDeleteCueList, input)
}
