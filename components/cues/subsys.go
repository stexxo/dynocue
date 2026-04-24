// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cues

import (
	"errors"

	"github.com/stexxo/dynocue/components/cues/types"
	"github.com/stexxo/dynocue/components/system"
	"github.com/stexxo/dynocue/core"
	"github.com/stexxo/dynocue/core/logging"
	"github.com/stexxo/dynocue/core/messaging"
)

type Cueing struct {
	*core.SubsystemCore
	persistence *system.PersistenceManager

	model *types.CueingModel
}

func New(logger logging.Logger) *Cueing {
	p := &Cueing{model: types.NewCueingModel()}
	p.SubsystemCore = core.NewSubsystemCore("cueing", logger, p.onStart)
	return p
}

func (p *Cueing) onStart() error {
	pm, err := system.RegisterWithPersistence(p.Messenger(), p.Logger(), p.Name(), SaveRequestSubject, LoadRequestSubject)
	if err != nil {
		return err
	}

	p.persistence = pm

	err = errors.Join(

		// Persistence
		messaging.Reply[string, string](p.Messenger(), false, SaveRequestSubject, p.Save),
		messaging.Reply[string, string](p.Messenger(), false, LoadRequestSubject, p.Load),

		// Cue Lists
		messaging.Reply[CreateCueListRequest, CreateCueListResponse](p.Messenger(), true, CreateCueListRequestSubject, p.CreateCueList),
		messaging.Reply[EnumerateCueListsRequest, EnumerateCueListsResponse](p.Messenger(), true, EnumerateCueListsRequestSubject, p.EnumerateCueLists),
		messaging.Reply[GetCueListByNumberRequest, GetCueListByNumberResponse](p.Messenger(), true, GetCueListByNumberRequestSubject, p.GetCueListByNumber),
		messaging.Reply[GetCueListByIdRequest, GetCueListByIdResponse](p.Messenger(), true, GetCueListByIdRequestSubject, p.GetCueListById),
		messaging.Reply[UpdateCueListMetadataRequest, UpdateCueListMetadataResponse](p.Messenger(), true, UpdateCueListMetadataRequestSubject, p.UpdateCueListMetadata),
		messaging.Reply[RenumberCueListsRequest, RenumberCueListsResponse](p.Messenger(), true, RenumberCueListRequestSubject, p.RenumberCueList),
		messaging.Reply[DeleteCueListsRequest, DeleteCueListsResponse](p.Messenger(), true, DeleteCueListRequestSubject, p.DeleteCueList),

		// Cues
		messaging.Reply[CreateCueRequest, CreateCueResponse](p.Messenger(), true, CreateCueRequestSubject, p.CreateCue),
		messaging.Reply[EnumerateCuesRequest, EnumerateCuesResponse](p.Messenger(), true, EnumerateCuesRequestSubject, p.EnumerateCues),
		messaging.Reply[GetCueByNumberRequest, GetCueByNumberResponse](p.Messenger(), true, GetCueByNumberRequestSubject, p.GetCueByNumber),
		messaging.Reply[GetCueByIdRequest, GetCueByIdResponse](p.Messenger(), true, GetCueByIdRequestSubject, p.GetCueById),
		messaging.Reply[UpdateCueMetadataRequest, UpdateCueMetadataResponse](p.Messenger(), true, UpdateCueMetadataRequestSubject, p.UpdateCueMetadata),
		messaging.Reply[RenumberCueRequest, RenumberCueResponse](p.Messenger(), true, RenumberCueRequestSubject, p.RenumberCue),
		messaging.Reply[DeleteCueRequest, DeleteCueResponse](p.Messenger(), true, DeleteCueRequestSubject, p.DeleteCue),
	)

	return err
}

const SaveRequestSubject = "request.cueing.persistence.save"

func (p *Cueing) Save(sub string, in *string) (*string, error) {
	p.Logger().Debug("attempting to save contents of subsystem show to stores")

	err := p.persistence.WriteToObjectStore("model", &p.model)
	if err != nil {
		return nil, err
	}

	return new(""), nil
}

const LoadRequestSubject = "request.cueing.persistence.load"
const LoadNotifyEventSubject = "event.cueing.persistence.loaded"

func (p *Cueing) Load(sub string, in *string) (*string, error) {
	p.Logger().Debug("attempting to load contents of subsystem cueing to stores")
	model := types.NewCueingModel()
	err := p.persistence.ReadFromObjectStore("model", model)
	if err != nil {
		return nil, err
	}
	p.model = model
	err = messaging.Publish(p.Messenger(), LoadNotifyEventSubject, "")
	if err != nil {
		return nil, err
	}
	return new(""), nil
}
