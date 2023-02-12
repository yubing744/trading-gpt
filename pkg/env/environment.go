package env

type Environment struct {
	entites []Entity
}

func NewEnvironment() *Environment {
	return &Environment{
		entites: make([]Entity, 0),
	}
}
