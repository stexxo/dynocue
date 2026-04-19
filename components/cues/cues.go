package cues

import (
	"github.com/stexxo/dynocue/components/system"
	"github.com/stexxo/dynocue/core"
	"github.com/stexxo/dynocue/core/logging"
	"github.com/stexxo/dynocue/core/messaging"
)

type Cueing struct {
	*core.SubsystemCore
	persistence *system.PersistenceManager

	model CueingModel
}

func New(logger logging.Logger) *Cueing {
	p := &Cueing{}
	p.SubsystemCore = core.NewSubsystemCore("cueing", logger, p.onStart)
	return p
}

func (p *Cueing) onStart() error {
	pm, err := system.RegisterWithPersistence(p.Messenger(), p.Logger(), p.Name(), SaveEventSubject, LoadEventSubject)
	if err != nil {
		return err
	}

	p.persistence = pm

	err = messaging.Reply[string, string](p.Messenger(), false, SaveEventSubject, p.Save)
	if err != nil {
		return err
	}

	err = messaging.Reply[string, string](p.Messenger(), false, LoadEventSubject, p.Load)
	if err != nil {
		return err
	}

	return nil
}

const SaveEventSubject = "request.cueing.persistence.save"

func (p *Cueing) Save(sub string, in *string) (*string, error) {
	p.Logger().Debug("attempting to save contents of subsystem show to stores")

	err := p.persistence.WriteToObjectStore("model", &p.model)
	if err != nil {
		return nil, err
	}

	return new(""), nil
}

const LoadEventSubject = "request.cueing.persistence.load"

func (p *Cueing) Load(sub string, in *string) (*string, error) {
	p.Logger().Debug("attempting to load contents of subsystem cueing to stores")
	err := p.persistence.ReadFromObjectStore("model", &p.model)
	if err != nil {
		return nil, err
	}
	return new(""), nil
}
