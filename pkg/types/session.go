package types

type ISession interface {
	GetID() string
	GetChats() []string
	AddChat(chat string)
}
