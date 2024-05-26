package types

import "github.com/google/uuid"

type IEvent interface {
	GetID() string
	GetType() string
	GetData() interface{}
	ToPrompts() []string
}

type Event struct {
	id    string
	ttype string
	data  interface{}
}

func NewEvent(ty string, data interface{}) *Event {
	return &Event{
		id:    uuid.NewString(),
		ttype: ty,
		data:  data,
	}
}

func (e *Event) GetID() string {
	return e.id
}

func (e *Event) GetType() string {
	return e.ttype
}

func (e *Event) GetData() interface{} {
	return e.data
}

func (e *Event) ToPrompts() []string {
	return []string{}
}
