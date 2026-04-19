package show

import (
	"github.com/stexxo/dynocue/components/system"
	"github.com/stexxo/dynocue/core"
	"github.com/stexxo/dynocue/core/logging"
	"github.com/stexxo/dynocue/core/messaging"
)

type Show struct {
	*core.SubsystemCore
	persistence *system.PersistenceManager
}

func New(logger logging.Logger) *Show {
	p := &Show{}
	p.SubsystemCore = core.NewSubsystemCore("show", logger, p.onStart)
	return p
}

func (p *Show) onStart() error {
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

const SaveEventSubject = "request.show.persistence.save"
const LoadEventSubject = "request.show.persistence.load"

func (p *Show) Save(sub string, in *string) (*string, error) {
	p.Logger().Debug("attempting to save contents of subsystem show to stores")
	return new(""), nil
}

func (p *Show) Load(sub string, in *string) (*string, error) {
	p.Logger().Debug("attempting to load contents of subsystem show from stores")
	return new(""), nil
}
