package types

import "context"

type IChannel interface {
	GetID() string
	OnMessage(cb MessageCallback)
	Reply(ctx context.Context, msg *Message) error
}
