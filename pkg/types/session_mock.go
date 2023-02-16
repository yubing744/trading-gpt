package types

type MockSession struct {
	id    string
	chats []string
}

func NewMockSession(id string) *MockSession {
	return &MockSession{
		id:    id,
		chats: make([]string, 0),
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
