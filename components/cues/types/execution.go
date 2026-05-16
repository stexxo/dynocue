package types

import "time"

type CueExecution struct {
	CueListId string `msgpack:"cueListId" json:"cueListId"`
	CueId     string `msgpack:"cueId" json:"cueId"`

	Active   bool `msgpack:"active" json:"active"`
	Selected bool `msgpack:"selected" json:"selected"`

	Elapsed time.Duration `msgpack:"elapsed" json:"elapsed"`

	DelayActive   bool          `msgpack:"delayActive" json:"delayActive"`
	DelayProgress time.Duration `msgpack:"delayProgress" json:"delayProgress"`

	FollowActive   bool          `msgpack:"followActive" json:"followActive"`
	FollowProgress time.Duration `msgpack:"followProgress" json:"followProgress"`
}

type ActionExecution struct {
	CueListId string `msgpack:"cueListId" json:"cueListId"`
	CueId     string `msgpack:"cueId" json:"cueId"`
	ActionId  string `msgpack:"actionId" json:"actionId"`

	Elapsed time.Duration `msgpack:"elapsed" json:"elapsed"`

	DelayActive   bool          `msgpack:"delayActive" json:"delayActive"`
	DelayProgress time.Duration `msgpack:"delayProgress" json:"delayProgress"`
}
