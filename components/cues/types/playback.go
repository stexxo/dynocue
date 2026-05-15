package types

type CueListSelectedCue struct {
	CueListId     string `msgpack:"cueListId" json:"cueListId"`
	SelectedCueId string `msgpack:"selectedCueId" json:"selectedCueId"`
}
