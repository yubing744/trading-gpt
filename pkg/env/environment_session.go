package env

import (
	"context"

	"github.com/kataras/go-events"
	"github.com/yubing744/trading-gpt/pkg/types"
)

type EnvironmentSession struct {
	events.EventEmmiter
	id    string
	chats []string
	state interface{}
}

func NewEnvironmentSession(id string) *EnvironmentSession {
	return &EnvironmentSession{
		EventEmmiter: events.New(),
		id:           id,
		chats:        make([]string, 0),
	}
}

func (s *EnvironmentSession) GetID() string {
	return s.id
}

func (s *EnvironmentSession) GetChats() []string {
	return s.chats
}

func (s *EnvironmentSession) AddChat(chat string) {
	s.chats = append(s.chats, chat)
}

func (s *EnvironmentSession) SetState(state interface{}) {
	s.state = state
}

func (s *EnvironmentSession) GetState() interface{} {
	return s.state
}

func (s *EnvironmentSession) Reply(ctx context.Context, msg *types.Message) error {
	s.Emit("reply", msg)
	return nil
}
