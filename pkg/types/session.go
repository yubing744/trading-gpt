package types

import "context"

type ISession interface {
	GetID() string
	GetChats() []string
	AddChat(chat string)
	Reply(ctx context.Context, msg *Message) error
	GetAttributeNames() []string
	GetAttribute(name string) (interface{}, bool)
	SetAttribute(name string, value interface{})
	RemoveAttribute(name string)
	SetRoles(role []string)
	HasRole(role string) bool
}
