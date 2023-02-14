package openai

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yubing744/trading-bot/pkg/config"
)

func TestNewOpenAIAgent(t *testing.T) {
	cfg := &config.AgentOpenAIConfig{
		Token: "xxxx",
	}
	agent := NewOpenAIAgent(cfg)
	assert.NotNil(t, agent)
}
