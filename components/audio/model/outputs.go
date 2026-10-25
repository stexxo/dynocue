package model

import (
	"github.com/google/uuid"
	"github.com/hashicorp/go-memdb"
	"github.com/stexxo/dynocue/components/audio/types"
	"github.com/stexxo/dynocue/db"
)

func (m *AudioModel) CreateAudioOutputWithDeviceName(number uint, label string, deviceName string) (uint, error) {
	return m.createAudioOutput(number, label, deviceName, false)
}

func (m *AudioModel) CreateAudioOutputWithSystemDefault(number uint, label string) (uint, error) {
	return m.createAudioOutput(number, label, "", true)
}

func (m *AudioModel) createAudioOutput(number uint, label string, deviceName string, useSystemDefault bool) (uint, error) {
	m.dbMu.RLock()
	defer m.dbMu.RUnlock()

	audioOutput := &types.AudioOutput{
		Id:               uuid.NewString(),
		OutputLabel:      label,
		AudioDeviceName:  deviceName,
		UseSystemDefault: useSystemDefault,
	}

	err := db.WithWrite(m.persistent, func(txn *memdb.Txn) error {
		num, err := getNextOutputNumber(txn, number)
		if err != nil {
			return err
		}
		audioOutput.Number = num
		return txn.Insert(TableDevices, audioOutput)
	})

	m.registry.Emit(ResourceOutput, OperationCreated, MetadataOutputId, audioOutput.Id)

	return audioOutput.Number, err
}

func getNextOutputNumber(txn *memdb.Txn, number uint) (uint, error) {
	if number == 0 {
		last, err := db.GetLastTxn[types.AudioOutput](txn, TableDevices, IndexNumber)
		if err != nil {
			return 0, err
		}
		return last.Number + 1, nil
	}

	existing, err := txn.First(TableDevices, IndexNumber, number)
	if err != nil {
		return 0, err
	}
	if existing != nil {
		return 0, ErrNumberExists
	}

	return number, nil
}

func (m *AudioModel) GetAudioOutput(outputId string) (*types.AudioOutput, error) {
	m.dbMu.RLock()
	defer m.dbMu.RUnlock()
	return db.GetFirstDb[types.AudioOutput](m.persistent, TableDevices, IndexId, outputId)
}

func (m *AudioModel) GetAudioOutputByNumber(number uint) (*types.AudioOutput, error) {
	m.dbMu.RLock()
	defer m.dbMu.RUnlock()
	return db.GetFirstDb[types.AudioOutput](m.persistent, TableDevices, IndexNumber, number)
}

func (m *AudioModel) EnumerateAudioOutputs() ([]types.AudioOutput, error) {
	m.dbMu.RLock()
	defer m.dbMu.RUnlock()
	return db.GetAllDb[types.AudioOutput](m.persistent, TableDevices, IndexNumber)
}

func (m *AudioModel) UpdateAudioOutput(outputId string, field string, value any) error {
	m.dbMu.RLock()
	defer m.dbMu.RUnlock()
	_, err := db.UpdateStructDb[types.AudioOutput](m.persistent, TableDevices, IndexId, outputId, field, value)
	if err != nil {
		return err
	}
	m.registry.Emit(ResourceOutput, OperationUpdated, MetadataOutputId, outputId)
	return nil
}

func (m *AudioModel) DeleteAudioOutput(outputId string) error {
	m.dbMu.RLock()
	defer m.dbMu.RUnlock()
	err := db.DeleteItemFromDb[types.AudioOutput](m.persistent, TableDevices, IndexId, outputId)
	if err != nil {
		return err
	}
	m.registry.Emit(ResourceOutput, OperationDeleted, MetadataOutputId, outputId)
	return nil
}
