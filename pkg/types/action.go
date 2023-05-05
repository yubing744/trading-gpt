package types

import (
	"encoding/json"
	"fmt"
	"strings"
)

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
	Name string            `json:"name"`
	Args map[string]string `json:"args"`
}

func (ac ActionDesc) ArgNames() []string {
	rets := make([]string, 0)

	for _, arg := range ac.Args {
		rets = append(rets, arg.Name)
	}

	return rets
}

func (ac ActionDesc) String() string {
	var argsText strings.Builder

	for i, arg := range ac.Args {
		argsText.WriteString("\"")
		argsText.WriteString(arg.Name)
		argsText.WriteString("\"")

		argsText.WriteString(": ")

		argsText.WriteString("\"<")
		argsText.WriteString(arg.Description)
		argsText.WriteString(">\"")

		if i < len(ac.Args)-1 {
			argsText.WriteString(",")
		}
	}

	if len(ac.Args) > 0 {
		return fmt.Sprintf(`%s: "%s", args: %s`, ac.Description, ac.Name, argsText.String())
	} else {
		return fmt.Sprintf(`%s: "%s"`, ac.Description, ac.Name)
	}
}

func (a *Action) JSON() string {
	data, err := json.Marshal(a)
	if err != nil {
		return "{}"
	}

	return string(data)
}
