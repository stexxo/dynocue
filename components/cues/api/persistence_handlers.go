package api

import (
	"errors"
	"io"

	"github.com/stexxo/dynocue/core/messaging"
)

func (c *CueingApi) registerPersistenceApis() error {
	return errors.Join(
		messaging.Reply[string, string](c.messenger, false, SaveRequestSubject, c.SaveModel),
		messaging.Reply[string, string](c.messenger, false, LoadRequestSubject, c.LoadModel),
	)
}

const SaveRequestSubject = "request.cueing.persistence.save"

const LoadRequestSubject = "request.cueing.persistence.load"

func (c *CueingApi) SaveModel(string, *string) (*string, error) {
	err := c.model.SerializeEachTable(func(name string, reader io.Reader) error {
		return c.persistence.WriteToObjectStore(name, reader)
	})

	return new(""), err
}

func (c *CueingApi) LoadModel(string, *string) (*string, error) {
	err := c.model.LoadModel(func(name string) (io.Reader, error) {
		return c.persistence.ReadFromObjectStore(name)
	})

	return new(""), err
}
