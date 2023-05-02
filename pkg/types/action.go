package types

import "encoding/json"

type ArgmentDesc struct {
	Name        string
	Description string
}

type Sample struct {
	Input  []string
	Output []string
}

type ActionDesc struct {
	Name        string
	Description string
	Args        []ArgmentDesc
	Samples     []Sample
}

type Action struct {
	Command string   `json:"cmd"`
	Args    []string `json:"args"`
}

func (ac ActionDesc) ArgNames() []string {
	rets := make([]string, 0)

	for _, arg := range ac.Args {
		rets = append(rets, arg.Name)
	}

	return rets
}

func (a *Action) JSON() string {
	data, err := json.Marshal(a)
	if err != nil {
		return "{}"
	}

	return string(data)
}
