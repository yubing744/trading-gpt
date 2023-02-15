package types

type ActionDesc struct {
	Name        string
	Description string
	Samples     []string
}

type Action struct {
	Name string
	Args []string
}
