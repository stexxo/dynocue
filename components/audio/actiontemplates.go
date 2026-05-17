// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package audio

import (
	"github.com/stexxo/dynocue/components/cues/api"
	"github.com/stexxo/dynocue/components/cues/types"
)

var (
	PlayActionTemplate = api.RegisterActionTemplateRequest{
		Template: types.ActionTemplate{
			TemplateId:    "action-play-v1",
			TemplateName:  "Play",
			SubsystemName: "Audio",
			Subject:       "action.audio.play",
			WaitForFinish: false,
			Fields: []types.ActionTemplateField{
				{FieldName: "source", FieldLabel: "Source", DataType: "float", DefaultValue: 0.0},
			},
		},
	}

	FadeActionTemplate = api.RegisterActionTemplateRequest{
		Template: types.ActionTemplate{
			TemplateId:    "action-fade-v1",
			TemplateName:  "Fade",
			SubsystemName: "Audio",
			Subject:       "action.audio.fade",
			WaitForFinish: true,
			Fields: []types.ActionTemplateField{
				{FieldName: "source", FieldLabel: "Source", DataType: "float", DefaultValue: 0.0},
				{FieldName: "targetVolume", FieldLabel: "Target Volume %", DataType: "float", DefaultValue: 0.0},
				{FieldName: "duration", FieldLabel: "Duration", DataType: "time", DefaultValue: 0.0},
			},
		},
	}

	StopActionTemplate = api.RegisterActionTemplateRequest{
		Template: types.ActionTemplate{
			TemplateId:    "action-stop-v1",
			TemplateName:  "Stop",
			SubsystemName: "Audio",
			Subject:       "action.audio.stop",
			WaitForFinish: false,
			Fields: []types.ActionTemplateField{
				{FieldName: "source", FieldLabel: "Source", DataType: "float", DefaultValue: 0.0},
			},
		},
	}

	PauseActionTemplate = api.RegisterActionTemplateRequest{
		Template: types.ActionTemplate{
			TemplateId:    "action-pause-v1",
			TemplateName:  "Pause",
			SubsystemName: "Audio",
			Subject:       "action.audio.pause",
			WaitForFinish: false,
			Fields: []types.ActionTemplateField{
				{FieldName: "source", FieldLabel: "Source", DataType: "float", DefaultValue: 0.0},
			},
		},
	}
)
