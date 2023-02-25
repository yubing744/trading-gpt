package types

import "context"

type INotifyChannel interface {
	GetID() string
	Reply(ctx context.Context, msg *Message) error
}
