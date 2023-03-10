package chat

import "github.com/yubing744/trading-gpt/pkg/types"

type ListenCallback func(ch types.IChannel)

type ChatProvider interface {
	GetName() string
	Listen(cb ListenCallback) error
}
