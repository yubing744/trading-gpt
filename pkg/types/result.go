package types

type Thoughts struct {
	Plan       string `json:"plan"`
	Analyze    string `json:"analyze"`
	Detail     string `json:"detail"`
	Reflection string `json:"reflection"`
	Speak      string `json:"speak"`
}

type Result struct {
	Thoughts *Thoughts `json:"thoughts"`
	Action   *Action   `json:"action"`
}
