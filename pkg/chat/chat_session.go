package chat

import (
	"context"

	"github.com/yubing744/trading-bot/pkg/types"
)

type ChatSession struct {
	id      string
	chats   []string
	roles   []string
	state   interface{}
	channel types.IChannel
}

func NewChatSession(channel types.IChannel) *ChatSession {
	return &ChatSession{
		id:      channel.GetID(),
		chats:   make([]string, 0),
		channel: channel,
	}
}

func (s *ChatSession) GetID() string {
	return s.id
}

func (s *ChatSession) GetChats() []string {
	return s.chats
}

func (s *ChatSession) AddChat(chat string) {
	s.chats = append(s.chats, chat)
}

func (s *ChatSession) SetState(state interface{}) {
	s.state = state
}

func (s *ChatSession) GetState() interface{} {
	return s.state
}

func (s *ChatSession) Reply(ctx context.Context, msg *types.Message) error {
	return s.channel.Reply(ctx, msg)
}

func (s *ChatSession) SetRoles(roles []string) {
	s.roles = roles
}

func (s *ChatSession) HasRole(role string) bool {
	for _, r := range s.roles {
		if r == role {
			return true
		}
	}

	return false
}
