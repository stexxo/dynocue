package cues

type cueListMetadata struct {
	Number   float64 `msgpack:"number"`
	Label    string  `msgpack:"label"`
	ListType string  `msgpack:"listType"`
}
