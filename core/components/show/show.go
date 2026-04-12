package show

import (
	"context"
	"time"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/stexxo/dynocue/core/components"
	"github.com/stexxo/dynocue/core/components/system"
	"github.com/stexxo/dynocue/core/logging"
	"github.com/stexxo/dynocue/core/messaging"
)

type Show struct {
	*components.BaseComponent
	kvStore     jetstream.KeyValue
	objectStore jetstream.ObjectStore
}

func NewShow(logger logging.Logger) *Show {
	p := &Show{}
	p.BaseComponent = components.NewBaseComponent("show", logger, p.onStart)
	return p
}

func (p *Show) onStart() error {

	p.Logger().Debug("attempting to register show subsystem with persistence")
	resp, err := messaging.RequestRetry[system.PersistenceRegistrationResponse](p.Messenger(), system.PersistenceRegistrationRequestSubject, system.PersistenceRegistrationRequest{SubsystemName: "show", SaveSubject: "request.show.persistence.save"}, 3, 100*time.Millisecond)
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

	go func() {
		time.Sleep(2 * time.Second)
		messaging.Request[system.PersistenceSaveResponse](p.Messenger(), system.PersistenceSaveRequestSubject, &system.PersistenceSaveRequest{Location: "/media/benjamin-dow/hermes/test.dyno"})
	}()

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
