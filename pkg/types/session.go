package types

import "context"

type ISession interface {
	GetID() string
	GetChats() []string
	AddChat(chat string)
	Reply(ctx context.Context, msg *Message) error
	SetState(state interface{})
	GetState() interface{}
}
