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

type CreateCueListInput struct {
	Number float64 `msgpack:"number" validate:"gte=0"`
}

type CreateCueListOutput struct {
	Number float64 `msgpack:"number" validate:"gte=0"`
}

type NewCueListEvent struct {
	Number float64 `msgpack:"number" validate:"gte=0"`
}

type UpdateCueListMetadataInput struct {
	Number float64 `msgpack:"number" validate:"gt=0"`
	Value  string  `msgpack:"value" validate:"required"`
}

type UpdateCueListMetadataOutput struct{}

type UpdateCueListMetadataEvent struct {
	Number float64 `msgpack:"number" validate:"gt=0"`
	Value  string  `msgpack:"value" validate:"required"`
}

type GetCueListMetadataInput struct {
	Number float64 `msgpack:"number" validate:"gt=0"`
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
	Number float64 `msgpack:"number" validate:"gt=0"`
}

type DeleteCueListOutput struct{}

type DeleteCueListEvent struct {
	Number float64 `msgpack:"number" validate:"gt=0"`
}
