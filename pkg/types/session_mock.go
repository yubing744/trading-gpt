package types

import (
	"context"
	"sync"
)

type MockSession struct {
	id         string
	chats      []string
	roles      []string
	attributes *sync.Map
}

func NewMockSession(id string) *MockSession {
	return &MockSession{
		id:         id,
		chats:      make([]string, 0),
		attributes: new(sync.Map),
	}
}

func (s *MockSession) GetID() string {
	return s.id
}

func (s *MockSession) GetChats() []string {
	return s.chats
}

func (s *MockSession) AddChat(chat string) {
	s.chats = append(s.chats, chat)
}

func (s *MockSession) GetAttributeNames() []string {
	names := make([]string, 0)

	s.attributes.Range(func(key, value any) bool {
		names = append(names, key.(string))
		return true
	})

	return names
}

func (s *MockSession) GetAttribute(name string) (interface{}, bool) {
	val, ok := s.attributes.Load(name)
	if ok {
		return val, true
	}

	return nil, false
}

func (s *MockSession) SetAttribute(name string, value interface{}) {
	s.attributes.Store(name, value)
}

func (s *MockSession) RemoveAttribute(name string) {
	s.attributes.Delete(name)
}

func (s *MockSession) Reply(ctx context.Context, msg *Message) error {
	return nil
}

func (s *MockSession) SetRoles(roles []string) {
	s.roles = roles
}

func (s *MockSession) HasRole(role string) bool {
	for _, r := range s.roles {
		if r == role {
			return true
		}
	}

	return false
}
