package api

import (
	"errors"
	"io"

	"github.com/stexxo/dynocue/core/messaging"
)

const (
	SaveRequestSubject = "request.audio.save"
	LoadRequestSubject = "request.audio.load"
)

func (a *AudioAPI) registerPersistenceApis() error {
	return errors.Join(
		messaging.Reply[string, string](a.messenger, false, SaveRequestSubject, a.SaveModel),
		messaging.Reply[string, string](a.messenger, false, LoadRequestSubject, a.LoadModel),
	)
}

func (a *AudioAPI) SaveModel(string, *string) (*string, error) {
	err := a.model.SerializeEachTable(func(name string, reader io.Reader) error {
		return a.persistence.WriteToObjectStore(name, reader)
	})

	return new(""), err
}

func (a *AudioAPI) LoadModel(string, *string) (*string, error) {
	err := a.model.LoadModel(func(name string) (io.Reader, error) {
		return a.persistence.ReadFromObjectStore(name)
	})

	return new(""), err
}
