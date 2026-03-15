package cues

const (
	RequestCreateCue         = "request.cue.create"
	RequestUpdateCueMetadata = "request.cue.metadata.update"
	RequestGetCueMetadata    = "request.cue.metadata.get"
	RequestEnumerateCue      = "request.cue.enumerate"
	RequestDeleteCue         = "request.cue.delete"
	RequestMoveCue           = "request.cue.move"

	EventNewCue    = "event.cue.created"
	EventUpdateCue = "event.cue.updated"
	EventDeleteCue = "event.cue.deleted"
)

type CreateCueInput struct {
	CueListNumber float64 `json:"cueListNumber" msgpack:"cueListNumber" validate:"gt=0"`
	Number        float64 `json:"number" msgpack:"number" validate:"gte=0"`
}

type CreateCueOutput struct {
	CueListNumber float64 `json:"cueListNumber" msgpack:"cueListNumber"`
	Number        float64 `json:"number" msgpack:"number"`
}

type NewCueEvent struct {
	CueListNumber float64 `json:"cueListNumber" msgpack:"cueListNumber"`
	Number        float64 `json:"number" msgpack:"number"`
	Label         string  `json:"label" msgpack:"label"`
}

type UpdateCueMetadataInput struct {
	CueListNumber float64 `json:"cueListNumber" msgpack:"cueListNumber" validate:"gt=0"`
	Number        float64 `json:"number" msgpack:"number" validate:"gt=0"`
	Key           string  `json:"key" msgpack:"key" validate:"required"`
	Value         string  `json:"value" msgpack:"value"`
}

type UpdateCueMetadataOutput struct{}

type UpdateCueMetadataEvent struct {
	CueListNumber float64 `json:"cueListNumber" msgpack:"cueListNumber"`
	Number        float64 `json:"number" msgpack:"number"`
	Label         string  `json:"label" msgpack:"label"`
}

type GetCueMetadataInput struct {
	CueListNumber float64 `json:"cueListNumber" msgpack:"cueListNumber" validate:"gt=0"`
	Number        float64 `json:"number" msgpack:"number" validate:"gt=0"`
}

type GetCueMetadataOutput struct {
	CueListNumber float64 `json:"cueListNumber" msgpack:"cueListNumber"`
	Number        float64 `json:"number" msgpack:"number"`
	Label         string  `json:"label" msgpack:"label"`
}

type EnumerateCueInput struct {
	CueListNumber float64 `json:"cueListNumber" msgpack:"cueListNumber" validate:"gt=0"`
}

type EnumerateCueOutput struct {
	CueListNumber float64                `json:"cueListNumber" msgpack:"cueListNumber"`
	Cues          []GetCueMetadataOutput `json:"cues" msgpack:"cues"`
}

type DeleteCueInput struct {
	CueListNumber float64 `json:"cueListNumber" msgpack:"cueListNumber" validate:"gt=0"`
	Number        float64 `json:"number" msgpack:"number" validate:"gt=0"`
}

type DeleteCueOutput struct{}

type DeleteCueEvent struct {
	CueListNumber float64 `json:"cueListNumber" msgpack:"cueListNumber"`
	Number        float64 `json:"number" msgpack:"number"`
}

type MoveCueInput struct {
	CueListNumber  float64 `json:"cueListNumber" msgpack:"cueListNumber" validate:"gt=0"`
	OriginalNumber float64 `json:"originalNumber" msgpack:"originalNumber" validate:"gt=0"`
	NewNumber      float64 `json:"newNumber" msgpack:"newNumber" validate:"gt=0"`
}

type MoveCueOutput struct {
	CueListNumber float64 `json:"cueListNumber" msgpack:"cueListNumber"`
	NewNumber     float64 `json:"newNumber" msgpack:"newNumber"`
}
