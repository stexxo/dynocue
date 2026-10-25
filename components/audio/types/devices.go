package types

type AudioDevice struct {
	Name         string `msgpack:"name" json:"name"`
	FriendlyName string `msgpack:"friendlyName" json:"friendlyName"`
	Description  string `msgpack:"description" json:"description"`
}

type AudioOutput struct {
	Id               string `msgpack:"outputId" json:"outputId"`
	Number           uint   `msgpack:"outputNumber" json:"outputNumber"`
	OutputLabel      string `msgpack:"outputLabel" json:"outputLabel"`
	AudioDeviceName  string `msgpack:"audioDeviceName" json:"audioDeviceName"`
	UseSystemDefault bool   `msgpack:"useSystemDefault" json:"useSystemDefault"`
}
