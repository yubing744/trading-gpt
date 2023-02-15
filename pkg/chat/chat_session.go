package chat

import (
	"github.com/kataras/go-events"
	"github.com/yubing744/trading-bot/pkg/agent"
	"github.com/yubing744/trading-bot/pkg/env"
	"github.com/yubing744/trading-bot/pkg/types"
)

type ChatSession struct {
	events.EventEmmiter

	Channel types.Channel
	Agent   agent.Agent
	Env     *env.Environment
}

func NewChatSession(channel types.Channel, agent agent.Agent, env *env.Environment) *ChatSession {
	return &ChatSession{
		EventEmmiter: events.New(),
		Channel:      channel,
		Agent:        agent,
		Env:          env,
	}
}
