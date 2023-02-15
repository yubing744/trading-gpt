package types

type ActionDesc struct {
	Name        string
	Description string
	Samples     []string
}

type Action struct {
	Target string
	Name   string
	Args   []string
}
