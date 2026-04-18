package show

import (
	"context"
	"time"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/stexxo/dynocue/components/system"
	"github.com/stexxo/dynocue/core"
	"github.com/stexxo/dynocue/core/logging"
	"github.com/stexxo/dynocue/core/messaging"
)

type Show struct {
	*core.SubsystemCore
	kvStore     jetstream.KeyValue
	objectStore jetstream.ObjectStore
}

func NewShow(logger logging.Logger) *Show {
	p := &Show{}
	p.SubsystemCore = core.NewSubsystemCore("show", logger, p.onStart)
	return p
}

func (p *Show) onStart() error {

	p.Logger().Debug("attempting to register show subsystem with persistence")
	resp, err := messaging.RequestRetry[system.PersistenceRegistrationResponse](p.Messenger(), system.PersistenceRegistrationRequestSubject, system.PersistenceRegistrationRequest{SubsystemName: "show", SaveSubject: "request.show.persistence.save"}, 10, 500*time.Millisecond)
	if err != nil {
		return err
	}

	p.kvStore, err = p.Messenger().JetStream().KeyValue(context.Background(), resp.Response.KeyValueStoreName)
	if err != nil {
		return err
	}

	p.objectStore, err = p.Messenger().JetStream().ObjectStore(context.Background(), resp.Response.ObjectStoreName)
	if err != nil {
		return err
	}

	err = messaging.Reply[string, string](p.Messenger(), false, "request.show.persistence.save", p.Save)
	if err != nil {
		return err
	}

	return nil
}

func (p *Show) Save(sub string, in *string) (*string, error) {
	p.Logger().Debug("attempting to save contents of subsystem show to stores")
	_, err := p.kvStore.Put(context.Background(), "show", []byte("test"))
	if err != nil {
		p.Logger().Error("failed to save contents of subsystem show to stores")
		return nil, err
	}

	return nil, nil
}
