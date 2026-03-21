// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package gui

import (
	"github.com/stexxo/dynocue/api/cues"
)

func (c *Commands) CreateCueList(input cues.CreateCueListInput) (*cues.CreateCueListOutput, error) {
	return makeRequest[cues.CreateCueListInput, cues.CreateCueListOutput](c, cues.RequestCreateCueList, input)
}

func (c *Commands) UpdateCueList(input cues.UpdateCueListInput) (*cues.UpdateCueListOutput, error) {
	return makeRequest[cues.UpdateCueListInput, cues.UpdateCueListOutput](c, cues.RequestUpdateCueList, input)
}

func (c *Commands) GetCueList(input cues.GetCueListInput) (*cues.GetCueListOutput, error) {
	return makeRequest[cues.GetCueListInput, cues.GetCueListOutput](c, cues.RequestGetCueList, input)
}

func (c *Commands) EnumerateCueList(input cues.EnumerateCueListInput) (*cues.EnumerateCueListOutput, error) {
	return makeRequest[cues.EnumerateCueListInput, cues.EnumerateCueListOutput](c, cues.RequestEnumerateCueList, input)
}

func (c *Commands) DeleteCueList(input cues.DeleteCueListInput) (*cues.DeleteCueListOutput, error) {
	return makeRequest[cues.DeleteCueListInput, cues.DeleteCueListOutput](c, cues.RequestDeleteCueList, input)
}

func (c *Commands) MoveCueList(input cues.MoveCueListInput) (*cues.MoveCueListOutput, error) {
	return makeRequest[cues.MoveCueListInput, cues.MoveCueListOutput](c, cues.RequestMoveCueList, input)
}
