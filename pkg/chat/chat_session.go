package chat

import (
	"context"
	"sync"

	"github.com/yubing744/trading-gpt/pkg/types"
)

type ChatSession struct {
	id         string
	chats      []string
	roles      []string
	attributes *sync.Map
	channel    types.INotifyChannel
}

func NewChatSession(channel types.INotifyChannel) *ChatSession {
	return &ChatSession{
		id:         channel.GetID(),
		chats:      make([]string, 0),
		attributes: new(sync.Map),
		channel:    channel,
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

func (s *ChatSession) GetAttributeNames() []string {
	names := make([]string, 0)

	s.attributes.Range(func(key, value any) bool {
		names = append(names, key.(string))
		return true
	})

	return names
}

func (s *ChatSession) GetAttribute(name string) (interface{}, bool) {
	val, ok := s.attributes.Load(name)
	if ok {
		return val, true
	}

	return nil, false
}

func (s *ChatSession) SetAttribute(name string, value interface{}) {
	s.attributes.Store(name, value)
}

func (s *ChatSession) RemoveAttribute(name string) {
	s.attributes.Delete(name)
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
