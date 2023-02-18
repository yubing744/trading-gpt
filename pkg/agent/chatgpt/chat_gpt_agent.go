package chatgpt

import (
	"github.com/sirupsen/logrus"
	"github.com/yubing744/trading-bot/pkg/config"
)

const (
	HumanLable = "You"
)

var log = logrus.WithField("agent", "chatgpt")

type ChatGPTAgent struct {
}

func NewOpenAIAgent(cfg *config.AgentOpenAIConfig) *ChatGPTAgent {
	return &ChatGPTAgent{}
}
