// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package engine

import (
	"errors"
	"time"

	"github.com/stexxo/dynocue/components/cues/model"
	"github.com/stexxo/dynocue/components/cues/types"
)

func (c *CueingEngine) GoToCue(cueId string) error {
	cue, err := c.model.GetCueById(cueId)
	if err != nil {
		return err
	}

	return c.goToCue(cue)
}

func (c *CueingEngine) GoToNextCue(cuelistId string) error {
	selected, err := c.model.GetSelectedCue(cuelistId)

	if err != nil && !errors.Is(err, model.ErrCueNotFound) {
		return err
	}

	var cue *types.Cue
	if errors.Is(err, model.ErrCueNotFound) {
		cue, err = c.model.GetFirstCueInCueList(cuelistId)
		if err != nil {
			return err
		}
	} else {
		cue, err = c.model.GetNextCueInCueList(cuelistId, selected.CueId)
		if err != nil {
			return err
		}
	}

	return c.goToCue(cue)
}

func (c *CueingEngine) goToCue(cue *types.Cue) error {
	// Get Cue List
	cueList, err := c.model.GetCueListById(cue.CueListId)
	if err != nil {
		return err
	}
	selectCue := cueList.CueListType == types.CueListTypeSequential

	err = c.model.StartCueExecution(cue.CueId, selectCue, true) // TODO make active when there is an engine to handle that
	if err != nil {
		return err
	}

	go func() {
		time.Sleep(time.Second * 5)
		c.model.StopCueExecution(cue.CueId)
	}()

	// TODO - Trigger Cue Playback
	return nil
}
