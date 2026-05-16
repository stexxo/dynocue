// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cues

import (
	"github.com/stexxo/dynocue/components/cues/api"
	"github.com/stexxo/dynocue/components/cues/engine"
	"github.com/stexxo/dynocue/components/cues/model"
	"github.com/stexxo/dynocue/components/system"
	"github.com/stexxo/dynocue/core"
	"github.com/stexxo/dynocue/core/logging"
)

type Cueing struct {
	*core.SubsystemCore
	model  *model.CueingModel
	engine *engine.CueingEngine
	api    *api.CueingApi
}

func New(logger logging.Logger) *Cueing {
	p := &Cueing{}
	p.SubsystemCore = core.NewSubsystemCore("cueing", logger, p.onStart)
	return p
}

func (p *Cueing) onStart() error {
	pm, err := system.RegisterWithPersistence(p.Messenger(), p.Logger(), p.Name(), api.SaveRequestSubject, api.LoadRequestSubject)
	if err != nil {
		return err
	}

	m, err := model.NewCueingModel()
	if err != nil {
		return err
	}
	p.model = m

	eng := engine.NewCueingEngine(m, p.Logger(), p.Messenger())
	p.engine = eng

	a, err := api.NewCueingApi(m, eng, pm, p.Messenger(), p.Logger())
	if err != nil {
		return err
	}
	p.api = a

	return err
}
