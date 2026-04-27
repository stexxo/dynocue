// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package audio

import (
	"github.com/stexxo/dynocue/components/cues"
	"github.com/stexxo/dynocue/components/cues/types"
)

var (
	PlayActionTemplate = cues.RegisterActionTemplateRequest{
		Id:            "action-play-v1",
		Name:          "Play",
		SubsystemName: "Audio",
		Subject:       "action.audio.play",
		Fields: []types.ActionTemplateField{
			{FieldName: "source", FieldLabel: "Source", DataType: "float", DefaultValue: 0.0},
		},
	}

	FadeActionTemplate = cues.RegisterActionTemplateRequest{
		Id:            "action-fade-v1",
		Name:          "Fade",
		SubsystemName: "Audio",
		Subject:       "action.audio.fade",
		Fields: []types.ActionTemplateField{
			{FieldName: "source", FieldLabel: "Source", DataType: "float", DefaultValue: 0.0},
			{FieldName: "targetVolume", FieldLabel: "Target Volume %", DataType: "float", DefaultValue: 0.0},
			{FieldName: "duration", FieldLabel: "Duration", DataType: "time", DefaultValue: 0.0},
		},
	}

	StopActionTemplate = cues.RegisterActionTemplateRequest{
		Id:            "action-stop-v1",
		Name:          "Stop",
		SubsystemName: "Audio",
		Subject:       "action.audio.stop",
		Fields: []types.ActionTemplateField{
			{FieldName: "source", FieldLabel: "Source", DataType: "float", DefaultValue: 0.0},
		},
	}

	PauseActionTemplate = cues.RegisterActionTemplateRequest{
		Id:            "action-pause-v1",
		Name:          "Pause",
		SubsystemName: "Audio",
		Subject:       "action.audio.pause",
		Fields: []types.ActionTemplateField{
			{FieldName: "source", FieldLabel: "Source", DataType: "float", DefaultValue: 0.0},
		},
	}
)
