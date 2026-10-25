// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package audio

import (
	"time"

	"github.com/stexxo/dynocue/components/audio/api"
	"github.com/stexxo/dynocue/components/audio/model"
	cueingapi "github.com/stexxo/dynocue/components/cues/api"
	"github.com/stexxo/dynocue/components/system"
	"github.com/stexxo/dynocue/core"
	"github.com/stexxo/dynocue/core/logging"
	"github.com/stexxo/dynocue/core/messaging"
	"github.com/stexxo/dynocue/pkg/vlc"
)

type Audio struct {
	*core.SubsystemCore
	model *model.AudioModel
	api   *api.AudioAPI
}

func New(logger logging.Logger) *Audio {
	p := &Audio{}
	p.SubsystemCore = core.NewSubsystemCore("audio", logger, p.onStart)
	return p
}

func (a *Audio) onStart() error {
	_, err := messaging.RequestRetry[cueingapi.RegisterActionTemplateResponse](a.Messenger(), cueingapi.RegisterActionTemplateRequestSubject, PlayActionTemplate, 10, time.Second)
	if err != nil {
		return err
	}

	_, err = messaging.RequestRetry[cueingapi.RegisterActionTemplateResponse](a.Messenger(), cueingapi.RegisterActionTemplateRequestSubject, FadeActionTemplate, 10, time.Second)
	if err != nil {
		return err
	}

	_, err = messaging.RequestRetry[cueingapi.RegisterActionTemplateResponse](a.Messenger(), cueingapi.RegisterActionTemplateRequestSubject, StopActionTemplate, 10, time.Second)
	if err != nil {
		return err
	}

	_, err = messaging.RequestRetry[cueingapi.RegisterActionTemplateResponse](a.Messenger(), cueingapi.RegisterActionTemplateRequestSubject, PauseActionTemplate, 10, time.Second)
	if err != nil {
		return err
	}

	pm, err := system.RegisterWithPersistence(a.Messenger(), a.Logger(), a.Name(), api.SaveRequestSubject, api.LoadRequestSubject)
	if err != nil {
		return err
	}

	m, err := model.NewAudioModel()
	if err != nil {
		return err
	}
	a.model = m

	audioApi, err := api.NewAudioAPI(a.model, pm, a.Messenger(), a.Logger())
	if err != nil {
		return err
	}

	a.api = audioApi

	// Ensure VLC is Accessible
	err = vlc.Initialize()
	if err != nil {
		return err
	}

	return err
}
