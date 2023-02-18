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
