package cues

const (
	RequestCreateCueList = "cuelist.create"
	EventNewCueList      = "change.cuelist.new"
)

type NewCueListInput struct {
	Number float64 `msgpack:"number"`
}

type NewCueListOutput struct {
	Number float64 `msgpack:"number"`
}

type NewCueListEvent struct {
	Number float64 `msgpack:"number"`
}
