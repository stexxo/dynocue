// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package gui

import (
	"github.com/stexxo/dynocue/api/cues"
)

func (c *Commands) CreateCue(input cues.CreateCueInput) (*cues.CreateCueOutput, error) {
	return makeRequest[cues.CreateCueInput, cues.CreateCueOutput](c, cues.RequestCreateCue, input)
}

func (c *Commands) UpdateCue(input cues.UpdateCueInput) (*cues.UpdateCueOutput, error) {
	return makeRequest[cues.UpdateCueInput, cues.UpdateCueOutput](c, cues.RequestUpdateCue, input)
}

func (c *Commands) GetCue(input cues.GetCueInput) (*cues.GetCueOutput, error) {
	return makeRequest[cues.GetCueInput, cues.GetCueOutput](c, cues.RequestGetCue, input)
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
