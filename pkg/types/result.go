package types

type Thoughts struct {
	Text      string `json:"text"`
	Analyze   string `json:"analyze"`
	Criticism string `json:"criticism"`
	Speak     string `json:"speak"`
}

type Result struct {
	Thoughts *Thoughts `json:"thoughts"`
	Action   *Action   `json:"action"`
}
