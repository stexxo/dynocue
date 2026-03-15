package cues

const (
	RequestCreateCueList         = "request.cuelist.create"
	RequestUpdateCueListMetadata = "request.cuelist.metadata.update"
	RequestGetCueListMetadata    = "request.cuelist.metadata.get"
	RequestEnumerateCueList      = "request.cuelist.enumerate"
	RequestDeleteCueList         = "request.cuelist.delete"
	RequestMoveCueList           = "request.cuelist.move"

	EventNewCueList    = "event.cuelist.created"
	EventUpdateCueList = "event.cuelist.updated"
	EventDeleteCueList = "event.cuelist.deleted"
)

type CreateCueListInput struct {
	Number float64 `json:"number" msgpack:"number" validate:"gte=0"`
}

type CreateCueListOutput struct {
	Number float64 `json:"number" msgpack:"number" validate:"gte=0"`
}

type NewCueListEvent struct {
	Number   float64 `json:"number" msgpack:"number" validate:"gte=0"`
	Label    string  `json:"label" msgpack:"label"`
	ListType string  `json:"listType" msgpack:"listType"`
}

type UpdateCueListMetadataInput struct {
	Number float64 `json:"number" msgpack:"number" validate:"gt=0"`
	Key    string  `json:"key" msgpack:"key" validate:"required"`
	Value  string  `json:"value" msgpack:"value"`
}

type UpdateCueListMetadataOutput struct{}

type UpdateCueListMetadataEvent struct {
	Number   float64 `json:"number" msgpack:"number" validate:"gt=0"`
	Label    string  `json:"label" msgpack:"label"`
	ListType string  `json:"listType" msgpack:"listType"`
}

type GetCueListMetadataInput struct {
	Number float64 `json:"number" msgpack:"number" validate:"gt=0"`
}

type GetCueListMetadataOutput struct {
	Number   float64 `json:"number" msgpack:"number"`
	Label    string  `json:"label" msgpack:"label"`
	ListType string  `json:"listType" msgpack:"listType"`
}

type EnumerateCueListInput struct{}

type EnumerateCueListOutput struct {
	CueLists []GetCueListMetadataOutput `json:"cueLists" msgpack:"cueLists"`
}

type DeleteCueListInput struct {
	Number float64 `json:"number" msgpack:"number" validate:"gt=0"`
}

type DeleteCueListOutput struct{}

type MoveCueListInput struct {
	OriginalNumber float64 `json:"originalNumber" msgpack:"originalNumber" validate:"gt=0"`
	NewNumber      float64 `json:"newNumber" msgpack:"newNumber" validate:"gt=0"`
}

type MoveCueListOutput struct {
	OriginalNumber float64 `json:"originalNumber" msgpack:"originalNumber"`
	NewNumber      float64 `json:"newNumber" msgpack:"newNumber"`
}

type DeleteCueListEvent struct {
	Number float64 `json:"number" msgpack:"number" validate:"gt=0"`
}
