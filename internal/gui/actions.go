// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package gui

import (
	"github.com/stexxo/dynocue/api/cues"
)

func (c *Commands) CreateAction(input cues.CreateActionInput) (*cues.CreateActionOutput, error) {
	return makeRequest[cues.CreateActionInput, cues.CreateActionOutput](c, cues.RequestCreateAction, input)
}

func (c *Commands) UpdateAction(input cues.UpdateActionInput) (*cues.UpdateActionOutput, error) {
	return makeRequest[cues.UpdateActionInput, cues.UpdateActionOutput](c, cues.RequestUpdateAction, input)
}

func (c *Commands) GetAction(input cues.GetActionInput) (*cues.GetActionOutput, error) {
	return makeRequest[cues.GetActionInput, cues.GetActionOutput](c, cues.RequestGetAction, input)
}

func (c *Commands) EnumerateAction(input cues.EnumerateActionInput) (*cues.EnumerateActionOutput, error) {
	return makeRequest[cues.EnumerateActionInput, cues.EnumerateActionOutput](c, cues.RequestEnumerateAction, input)
}

func (c *Commands) DeleteAction(input cues.DeleteActionInput) (*cues.DeleteActionOutput, error) {
	return makeRequest[cues.DeleteActionInput, cues.DeleteActionOutput](c, cues.RequestDeleteAction, input)
}

func (c *Commands) MoveAction(input cues.MoveActionInput) (*cues.MoveActionOutput, error) {
	return makeRequest[cues.MoveActionInput, cues.MoveActionOutput](c, cues.RequestMoveAction, input)
}
