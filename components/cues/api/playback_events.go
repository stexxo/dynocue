package api

import (
	"github.com/stexxo/dynocue/components/cues/model"
	"github.com/stexxo/dynocue/util"
)

func (c *CueingApi) registerPlaybackEvents() {
	c.model.RegisterEventHandler(model.ResourceCueListCueSelection, model.OperationUpdated, eventHandler[SelectedCueChangedEvent](c.messenger, c.logger, c.SelectedCueChanged))
}

const SelectedCueChangedEventSubject = "event.cueing.playback.cue.selected"

type SelectedCueChangedEvent struct {
	CueListId string `msgpack:"cueListId" json:"cueListId"`
	CueId     string `msgpack:"cueId" json:"cueId"`
}

func (c *CueingApi) SelectedCueChanged(ev util.Event) (string, *SelectedCueChangedEvent) {
	return SelectedCueChangedEventSubject, &SelectedCueChangedEvent{CueListId: ev.EventData[model.MetadataCueListId], CueId: ev.EventData[model.MetadataCueId]}
}
