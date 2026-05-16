package engine

import "github.com/stexxo/dynocue/components/cues/types"

func (c *CueingEngine) GoToCue(cueId string) error {
	cue, err := c.model.GetCueById(cueId)
	if err != nil {
		return err
	}

	return c.goToCue(cue)
}

func (c *CueingEngine) GoToNextCue(cuelistId string) error {
	selected, err := c.model.GetSelectedCue(cuelistId)
	if err != nil {
		return err
	}

	cue, err := c.model.GetNextCueInCueList(cuelistId, selected)
	if err != nil {
		return err
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

	err = c.model.StartCueExecution(cue.CueId, selectCue, false) // TODO make active when there is an engine to handle that
	if err != nil {
		return err
	}

	// TODO - Trigger Cue Playback
	return nil
}
