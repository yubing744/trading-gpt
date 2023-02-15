package chat

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/kataras/go-events"
	"github.com/sirupsen/logrus"
	"github.com/yubing744/trading-bot/pkg/agent"
	"github.com/yubing744/trading-bot/pkg/env"
	"github.com/yubing744/trading-bot/pkg/types"
)

var log = logrus.WithField("chat", "ChatSession")

type ChatSession struct {
	events.EventEmmiter

	channel types.Channel
	agent   agent.Agent
	env     *env.Environment
}

func NewChatSession(channel types.Channel, agent agent.Agent, env *env.Environment) *ChatSession {
	return &ChatSession{
		EventEmmiter: events.New(),
		channel:      channel,
		agent:        agent,
		env:          env,
	}
}

func (session *ChatSession) Start() error {
	session.channel.OnMessage(session.handleChatMessage)
	session.env.OnEvent(session.handleEnvEvent)
	return nil
}

func (session *ChatSession) handleChatMessage(msg *types.Message) {
	log.WithField("msg", msg).Info("new message")

	ctx := context.Background()
	evt := &types.Event{
		ID:   msg.ID,
		Type: "text_message",
		Data: msg.Text,
	}

	result, err := session.agent.GenActions(ctx, uuid.NewString(), evt)
	if err != nil {
		log.WithError(err).Error("gen action error")
		return
	}

	log.WithField("result", result).Info("gen actions result")

	if len(result.Actions) > 0 {
		for _, action := range result.Actions {
			session.env.SendCommand(context.Background(), action.Target, action.Name, action.Args)
		}
	}

	if len(result.Texts) > 0 {
		err = session.channel.Reply(&types.Message{
			ID:   uuid.NewString(),
			Text: strings.Join(result.Texts, ""),
		})
		if err != nil {
			log.WithError(err).Error("reply message error")
		}

		return
	}

	err = session.channel.Reply(&types.Message{
		ID:   uuid.NewString(),
		Text: "no reply text",
	})
	if err != nil {
		log.WithError(err).Error("reply message error")
	}
}

func (session *ChatSession) handleEnvEvent(evt *types.Event) {
	log.WithField("event", evt).Info("handle env event")
}
