package types

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
	Target string
	Name   string
	Args   []string
}

func (ac ActionDesc) ArgNames() []string {
	rets := make([]string, 0)

	for _, arg := range ac.Args {
		rets = append(rets, arg.Name)
	}

	return rets
}
