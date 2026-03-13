package cues

const (
	RequestCreateCueList         = "request.cuelist.create"
	RequestUpdateCueListMetadata = "request.cuelist.metadata.update.*"
	RequestGetCueListMetadata    = "request.cuelist.metadata.get"
	RequestEnumerateCueList      = "request.cuelist.enumerate"
	RequestDeleteCueList         = "request.cuelist.delete"

	EventNewCueList    = "event.cuelist.created"
	EventUpdateCueList = "event.cuelist.updated"
	EventDeleteCueList = "event.cuelist.deleted"
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

type UpdateCueListMetadataInput struct {
	Number float64 `msgpack:"number"`
	Value  string  `msgpack:"value"`
}

type UpdateCueListMetadataOutput struct{}

type UpdateCueListMetadataEvent struct {
	Number float64 `msgpack:"number"`
	Value  string  `msgpack:"value"`
}

type GetCueListMetadataInput struct {
	Number float64 `msgpack:"number"`
}

type GetCueListMetadataOutput struct {
	Number   float64 `msgpack:"number"`
	Label    string  `msgpack:"label"`
	ListType string  `msgpack:"listType"`
}

type EnumerateCueListInput struct{}

type EnumerateCueListOutput struct {
	CueLists []struct {
		Number   float64 `msgpack:"number"`
		Label    string  `msgpack:"label"`
		ListType string  `msgpack:"listType"`
	} `msgpack:"cueLists"`
}

type DeleteCueListInput struct {
	Number float64 `msgpack:"number"`
}

type DeleteCueListOutput struct{}

type DeleteCueListEvent struct {
	Number float64 `msgpack:"number"`
}
