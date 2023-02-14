package chat

import "github.com/yubing744/trading-bot/pkg/types"

type ListenCallback func(ch types.Channel)

type ChatProvider interface {
	GetName() string
	Listen(cb ListenCallback) error
}
