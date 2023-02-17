package types

import "context"

type Channel interface {
	GetID() string
	OnMessage(cb MessageCallback)
	Reply(ctx context.Context, msg *Message) error
}
