// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package engine

import (
	"errors"

	"github.com/stexxo/dynocue/components/cues/model"
	"github.com/stexxo/dynocue/components/cues/types"
)

func (c *CueingEngine) GoToCue(cueId string) error {
	return c.ExecuteCue(cueId)
}

func (c *CueingEngine) GoToNextCue(cuelistId string) error {
	// Start with the selected cue
	selected, err := c.model.GetSelectedCue(cuelistId)
	if err != nil && !errors.Is(err, model.ErrCueNotFound) {
		return err
	}

	var cue *types.Cue
	if errors.Is(err, model.ErrCueNotFound) { // if the cue is not found, start with the first cue in the cue list
		cue, err = c.model.GetFirstCueInCueList(cuelistId)
		if err != nil {
			return err
		}
	} else { // otherwise, start with the next cue
		cue, err = c.model.GetNextCueInCueList(cuelistId, selected.CueId)
		if err != nil {
			return err
		}
	}

	return c.ExecuteCue(cue.CueId)
}
