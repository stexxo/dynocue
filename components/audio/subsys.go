// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package audio

import (
	"bytes"
	"encoding/json"
	"errors"
	"time"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/stexxo/dynocue/components/audio/types"
	"github.com/stexxo/dynocue/components/cues"
	"github.com/stexxo/dynocue/components/system"
	"github.com/stexxo/dynocue/core"
	"github.com/stexxo/dynocue/core/logging"
	"github.com/stexxo/dynocue/core/messaging"
)

type Audio struct {
	*core.SubsystemCore
	persistence *system.PersistenceManager

	model *types.AudioModel
}

func New(logger logging.Logger) *Audio {
	p := &Audio{}
	p.model = types.NewAudioModel()
	p.SubsystemCore = core.NewSubsystemCore("audio", logger, p.onStart)
	return p
}

func (a *Audio) onStart() error {
	pm, err := system.RegisterWithPersistence(a.Messenger(), a.Logger(), a.Name(), SaveRequestSubject, LoadRequestSubject)
	if err != nil {
		return err
	}

	_, err = messaging.RequestRetry[cues.RegisterActionTemplateResponse](a.Messenger(), cues.RegisterActionTemplateRequestSubject, PlayActionTemplate, 10, time.Second)
	if err != nil {
		return err
	}

	_, err = messaging.RequestRetry[cues.RegisterActionTemplateResponse](a.Messenger(), cues.RegisterActionTemplateRequestSubject, FadeActionTemplate, 10, time.Second)
	if err != nil {
		return err
	}

	_, err = messaging.RequestRetry[cues.RegisterActionTemplateResponse](a.Messenger(), cues.RegisterActionTemplateRequestSubject, StopActionTemplate, 10, time.Second)
	if err != nil {
		return err
	}

	_, err = messaging.RequestRetry[cues.RegisterActionTemplateResponse](a.Messenger(), cues.RegisterActionTemplateRequestSubject, PauseActionTemplate, 10, time.Second)
	if err != nil {
		return err
	}

	a.persistence = pm

	err = errors.Join(
		// Persistence
		messaging.Reply[string, string](a.Messenger(), false, SaveRequestSubject, a.Save),
		messaging.Reply[string, string](a.Messenger(), false, LoadRequestSubject, a.Load),
	)

	return err
}

const SaveRequestSubject = "request.cueing.persistence.save"

func (a *Audio) Save(sub string, in *string) (*string, error) {
	a.Logger().Debug("attempting to save contents of subsystem audio to stores")

	buf, err := json.Marshal(a.model)
	if err != nil {
		return nil, err
	}

	err = a.persistence.WriteToObjectStore("model", bytes.NewReader(buf))
	if err != nil {
		return nil, err
	}

	return new(""), nil
}

const LoadRequestSubject = "request.cueing.persistence.load"
const LoadNotifyEventSubject = "event.cueing.persistence.loaded"

func (a *Audio) Load(sub string, in *string) (*string, error) {
	a.Logger().Debug("attempting to load contents of subsystem audio to stores")

	buf, err := a.persistence.ReadFromObjectStore("model")
	if errors.Is(err, jetstream.ErrObjectNotFound) {
		return new(""), nil
	}
	if err != nil {
		return nil, err
	}

	model := types.NewAudioModel()
	err = json.Unmarshal(buf.Bytes(), model)
	if err != nil {
		return nil, err
	}

	a.model = model
	err = messaging.Publish(a.Messenger(), LoadNotifyEventSubject, "")
	if err != nil {
		return nil, err
	}
	return new(""), nil
}
