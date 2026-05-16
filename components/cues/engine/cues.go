// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package engine

import (
	"errors"
	"time"

	"github.com/stexxo/dynocue/components/cues/types"
)

func (c *CueingEngine) ExecuteCue(cueId string) error {
	// Start Cue In Model
	err := c.startCue(cueId)
	if err != nil {
		return err
	}

	go func() {
		err := c.executeCue(cueId)
		if err != nil {
			c.logger.Error("cue failed in execution", "err", err)
		}
	}()

	return nil
}

func (c *CueingEngine) executeCue(cueId string) (err error) {
	delayFinished := false
	actionsStarted := false
	actionsFinished := false
	followFinished := false
	ticker := time.NewTicker(time.Millisecond * 10)

	defer func() {
		ticker.Stop()
		derr := c.model.StopCueExecution(cueId)
		if derr != nil {
			err = errors.Join(err, derr)
		}
	}()

	for {
		select {
		case <-ticker.C:
			if !delayFinished {
				delayFinished, err = c.checkDelay(cueId)
				if err != nil {
					return err
				}
			}

			if delayFinished && !actionsStarted {
				err = c.startActions(cueId)
				if err != nil {
					return err
				}
				actionsStarted = true
			}

			if actionsStarted && !actionsFinished {
				actionsFinished, err = c.checkActions(cueId)
				if err != nil {
					return err
				}
			}

			if !followFinished {
				followFinished, err = c.checkFollow(cueId)
				if err != nil {
					return err
				}
			}
		}

		if delayFinished && actionsFinished && followFinished {
			return
		}
	}
}

func (c *CueingEngine) startCue(cueId string) error {
	// Get CueList and Cue
	cue, err := c.model.GetCueById(cueId)
	if err != nil {
		return err
	}

	cueList, err := c.model.GetCueListById(cue.CueListId)
	if err != nil {
		return err
	}

	// Start the Execution in the DB
	err = c.model.StartCueExecution(cue.CueId, cueList.CueListType == types.CueListTypeSequential, true)
	if err != nil {
		return err
	}

	return nil
}

func (c *CueingEngine) checkDelay(cueId string) (bool, error) {
	cue, err := c.model.GetCueById(cueId)
	if err != nil {
		return false, err
	}

	if cue.Delay == 0 {
		err := c.model.StopCueDelayExecution(cueId) // handle case where it was started and then delay was set to 0 after it begun
		if err != nil {
			return false, err
		}
		return true, nil
	}

	cueExec, err := c.model.GetCueExecution(cueId)
	if err != nil {
		return false, err
	}

	if !cueExec.DelayActive {
		err = c.model.StartCueDelayExecution(cueId)
		if err != nil {
			return false, err
		}
		return false, nil
	}

	if time.Since(cueExec.DelayStart) >= cue.Delay {
		err := c.model.StopCueDelayExecution(cueId)
		if err != nil {
			return false, err
		}
		return true, nil
	}

	return false, nil
}

func (c *CueingEngine) checkFollow(cueId string) (bool, error) {
	cue, err := c.model.GetCueById(cueId)
	if err != nil {
		return false, err
	}

	if cue.Follow == 0 {
		return true, nil
	}

	cueExec, err := c.model.GetCueExecution(cueId)
	if err != nil {
		return false, err
	}

	if !cueExec.FollowActive {
		err = c.model.StartCueFollowExecution(cueId, cue.Follow)
		if err != nil {
			return false, err
		}
		return false, nil
	}

	if time.Since(cueExec.FollowStart) >= cue.Follow {
		err := c.model.StopCueFollowExecution(cueId)
		if err != nil {
			return false, err
		}

		// Trigger the next cue if the current cue is still selected
		cue, err := c.model.GetCueById(cueId)
		if err != nil {
			return false, err
		}
		cueExec, err := c.model.GetCueExecution(cueId)
		if err != nil {
			return false, err
		}
		if cueExec.Selected {
			nextCue, err := c.model.GetNextCueInCueList(cue.CueListId, cueId)
			if err != nil {
				return false, err
			}

			err = c.GoToCue(nextCue.CueId)
			if err != nil {
				return false, err
			}
		}

		return true, nil
	}

	return false, nil
}

func (c *CueingEngine) startActions(cueId string) error {
	actions, err := c.model.EnumerateActions(cueId)
	if err != nil {
		return err
	}

	for _, action := range actions {
		err := c.ExecuteAction(action.ActionId)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *CueingEngine) checkActions(cueId string) (bool, error) {
	actionExec, err := c.model.EnumerateActionExecutions(cueId)
	if err != nil {
		return false, err
	}
	if len(actionExec) > 0 {
		return false, nil
	}
	return true, nil
}
