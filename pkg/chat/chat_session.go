package chat

import (
	"github.com/kataras/go-events"
	"github.com/yubing744/trading-bot/pkg/agent"
	"github.com/yubing744/trading-bot/pkg/env"
	"github.com/yubing744/trading-bot/pkg/types"
)

type ChatSession struct {
	events.EventEmmiter

	id    string
	chats []string
	state interface{}

	Channel types.Channel
	Agent   agent.Agent
	Env     *env.Environment
}

func NewChatSession(channel types.Channel, agent agent.Agent, env *env.Environment) *ChatSession {
	return &ChatSession{
		id:           channel.GetID(),
		chats:        make([]string, 0),
		EventEmmiter: events.New(),
		Channel:      channel,
		Agent:        agent,
		Env:          env,
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
