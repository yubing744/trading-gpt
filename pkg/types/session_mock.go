package types

type MockSession struct {
	id    string
	chats []string
	state interface{}
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

func (s *MockSession) SetState(state interface{}) {
	s.state = state
}

func (s *MockSession) GetState() interface{} {
	return s.state
}
