package types

type ArgmentDesc struct {
	Name        string
	Description string
}

type Sample struct {
}

type ActionDesc struct {
	Name        string
	Description string
	Args        []ArgmentDesc
	Samples     []string
}

type Action struct {
	Target string
	Name   string
	Args   []string
}
