// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package engine

import (
	"github.com/stexxo/dynocue/components/cues/model"
	"github.com/stexxo/dynocue/core/logging"
)

type CueingEngine struct {
	model  *model.CueingModel
	logger logging.Logger
}

func NewCueingEngine(m *model.CueingModel) *CueingEngine {
	return &CueingEngine{
		model: m,
	}
}
