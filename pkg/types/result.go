package types

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

type Thoughts struct {
	Plan       interface{} `json:"plan"`
	Analyze    interface{} `json:"analyze"`
	Detail     interface{} `json:"detail"`
	Reflection interface{} `json:"reflection"`
	Speak      string      `json:"speak"`
}

// toHumanText converts the Thoughts struct into a human-readable string.
func (t *Thoughts) ToHumanText() string {
	return fmt.Sprintf("Plan: %s\nAnalyze: %s\nDetail: %s\nReflection: %s\nSpeak: %s\n",
		interfaceToString(t.Plan),
		interfaceToString(t.Analyze),
		interfaceToString(t.Detail),
		interfaceToString(t.Reflection),
		t.Speak)
}

// interfaceToString converts an interface{} to a string in a human-readable format.
func interfaceToString(i interface{}) string {
	if i == nil {
		return "none"
	}
	switch v := i.(type) {
	case string:
		return v
	case []string:
		return strings.Join(v, ", ")
	case fmt.Stringer:
		return v.String()
	default:
		// Attempt to marshal the value to JSON for a more readable output
		b, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprintf("unknown (%s)", reflect.TypeOf(i))
		}
		return string(b)
	}
}

type Result struct {
	Thoughts *Thoughts `json:"thoughts"`
	Action   *Action   `json:"action"`
}
