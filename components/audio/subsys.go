// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package audio

import (
	"time"

	"github.com/stexxo/dynocue/components/cues/api"
	"github.com/stexxo/dynocue/core"
	"github.com/stexxo/dynocue/core/logging"
	"github.com/stexxo/dynocue/core/messaging"
)

type Audio struct {
	*core.SubsystemCore
}

func New(logger logging.Logger) *Audio {
	p := &Audio{}
	p.SubsystemCore = core.NewSubsystemCore("audio", logger, p.onStart)
	return p
}

func (a *Audio) onStart() error {
	_, err := messaging.RequestRetry[api.RegisterActionTemplateResponse](a.Messenger(), api.RegisterActionTemplateRequestSubject, PlayActionTemplate, 10, time.Second)
	if err != nil {
		return err
	}

	_, err = messaging.RequestRetry[api.RegisterActionTemplateResponse](a.Messenger(), api.RegisterActionTemplateRequestSubject, FadeActionTemplate, 10, time.Second)
	if err != nil {
		return err
	}

	_, err = messaging.RequestRetry[api.RegisterActionTemplateResponse](a.Messenger(), api.RegisterActionTemplateRequestSubject, StopActionTemplate, 10, time.Second)
	if err != nil {
		return err
	}

	_, err = messaging.RequestRetry[api.RegisterActionTemplateResponse](a.Messenger(), api.RegisterActionTemplateRequestSubject, PauseActionTemplate, 10, time.Second)
	if err != nil {
		return err
	}
	return err
}
