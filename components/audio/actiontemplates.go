package audio

import (
	"github.com/stexxo/dynocue/components/cues"
	"github.com/stexxo/dynocue/components/cues/types"
)

var (
	PlayActionTemplate cues.RegisterActionTemplateRequest = cues.RegisterActionTemplateRequest{
		Name:    "Play",
		Subject: "action.audio.play",
		Fields: []types.ActionTemplateField{
			{FieldName: "source", FieldLabel: "Source", DataType: "float"},
		},
	}
)
