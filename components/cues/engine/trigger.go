package engine

func (c *CueingEngine) GoToCue(cueId string) error {
	cue, err := c.model.GetCueById(cueId)
	if err != nil {
		return err
	}

	err = c.model.SetSelectedCueId(cue.CueListId, cue.CueId)
	if err != nil {
		return err
	}

	// TODO - Trigger Cue Playback

	return nil
}

func (c *CueingEngine) GoToNextCue(cuelistId string) error {
	cueList, err := c.model.GetCueListById(cuelistId)
	if err != nil {
		return err
	}

	selected, err := c.model.GetSelectedCueId(cueList.CueListId)
	if err != nil {
		return err
	}

	cue, err := c.model.GetNextCueInCueList(cuelistId, selected.SelectedCueId)
	if err != nil {
		return err
	}

	err = c.model.SetSelectedCueId(cue.CueListId, cue.CueId)
	if err != nil {
		return err
	}

	// TODO - Trigger Cue Playback

	return nil
}
