package openai

import (
	"github.com/yubing744/trading-bot/pkg/config"
	"github.com/yubing744/trading-bot/pkg/types"
)

type OpenAIAgent struct {
	token string
}

func NewOpenAIAgent(cfg *config.AgentOpenAIConfig) *OpenAIAgent {
	return &OpenAIAgent{
		token: cfg.Token,
	}
}

func (agent *OpenAIAgent) GenAction(sessionID string, event *types.Event) (*types.Action, error) {
	return nil, nil
}
