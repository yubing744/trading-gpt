package openai

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yubing744/trading-bot/pkg/config"
	"github.com/yubing744/trading-bot/pkg/types"
)

func TestNewOpenAIAgent(t *testing.T) {
	cfg := &config.AgentOpenAIConfig{
		Token: "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",
	}
	agent := NewOpenAIAgent(cfg)
	assert.NotNil(t, agent)
}

func TestGenAction(t *testing.T) {
	ctx := context.Background()

	cfg := &config.AgentOpenAIConfig{
		Token: "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",
	}
	agent := NewOpenAIAgent(cfg)
	assert.NotNil(t, agent)

	agent.SetBackgroup("以下是和股票交易助手的对话，股票交易助手支持注册实体，支持输出指令控制实体，支持根据股价数据生成K线形态。")
	agent.RegisterActions(ctx, []*types.ActionDesc{
		{
			Name:        "buy",
			Description: "购买指令",
			Samples: []string{
				"1.0 1.1 1.2 1.3 1.4 1.5 1.6",
				"1.0 1.1 1.0 1.1 1.2 1.3 1.4",
			},
		},
		{
			Name:        "sell",
			Description: "卖出指令",
			Samples: []string{
				"1.6 1.5 1.4 1.3 1.2 1.1 1.0",
				"1.0 1.1 1.2 1.3 1.4 1.3 1.2",
			},
		},
		{
			Name:        "hold",
			Description: "持仓",
			Samples: []string{
				"1.2 1.3 1.4 1.5 1.6 1.7 1.8",
			},
		},
	})

	result, err := agent.GenActions(ctx, "test_session_1", &types.Event{
		ID:   "1",
		Type: "text_message",
		Data: "1.1 1.2 1.3 1.4 1.5 1.6 1.7",
	})
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "buy", result.Actions[0].Name)
}
