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
		Token:            "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",
		MaxContextLength: 500,
	}
	agent := NewOpenAIAgent(cfg)
	assert.NotNil(t, agent)
}

func TestGenPrompt(t *testing.T) {
	cfg := &config.AgentOpenAIConfig{
		Token:            "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",
		MaxContextLength: 500,
	}
	agent := NewOpenAIAgent(cfg)
	assert.NotNil(t, agent)

	prompt, err := agent.GenPrompt([]string{}, &types.Event{
		Data: "xxx",
	})

	assert.NoError(t, err)
	assert.Equal(t, "\n\nYou:xxx\nAI:", prompt)
}

func TestGenPromptWithTooLong(t *testing.T) {
	cfg := &config.AgentOpenAIConfig{
		Token:            "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",
		MaxContextLength: 12,
	}
	agent := NewOpenAIAgent(cfg)
	assert.NotNil(t, agent)

	_, err := agent.GenPrompt([]string{}, &types.Event{
		Data: "xxx",
	})

	assert.Error(t, err)
}

func TestGenPromptBySplitChats(t *testing.T) {
	cfg := &config.AgentOpenAIConfig{
		Token:            "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",
		MaxContextLength: 13,
	}
	agent := NewOpenAIAgent(cfg)
	assert.NotNil(t, agent)

	prompt, err := agent.GenPrompt([]string{"msg xxxx"}, &types.Event{
		Data: "xxx",
	})

	assert.NoError(t, err)
	assert.Equal(t, "\n\nYou:xxx\nAI:", prompt)
}

func TestGenPromptBySplitChats2(t *testing.T) {
	cfg := &config.AgentOpenAIConfig{
		Token:            "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",
		MaxContextLength: 23,
	}
	agent := NewOpenAIAgent(cfg)
	assert.NotNil(t, agent)

	prompt, err := agent.GenPrompt([]string{"msg xxxx1", "msg xxxx2"}, &types.Event{
		Data: "xxx",
	})

	assert.NoError(t, err)
	assert.Equal(t, "\n\nmsg xxxx2\nYou:xxx\nAI:", prompt)
}

func TestGenPromptBySplitChats3(t *testing.T) {
	cfg := &config.AgentOpenAIConfig{
		Token:            "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",
		MaxContextLength: 100,
	}
	agent := NewOpenAIAgent(cfg)
	assert.NotNil(t, agent)

	prompt, err := agent.GenPrompt([]string{"msg xxxx1", "msg xxxx2"}, &types.Event{
		Data: "xxx",
	})

	assert.NoError(t, err)
	assert.Equal(t, "\n\nmsg xxxx1\nmsg xxxx2\nYou:xxx\nAI:", prompt)
}

func TestGenAction(t *testing.T) {
	ctx := context.Background()

	cfg := &config.AgentOpenAIConfig{
		Token:            "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",
		MaxContextLength: 500,
	}
	agent := NewOpenAIAgent(cfg)
	assert.NotNil(t, agent)

	agent.SetBackgroup("以下是和股票交易助手的对话，股票交易助手支持注册实体，支持输出指令控制实体，支持根据股价数据生成K线形态。")
	agent.RegisterActions(ctx, "exchange", []*types.ActionDesc{
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

	session := types.NewMockSession("session_1")

	result, err := agent.GenActions(ctx, session, &types.Event{
		ID:   "1",
		Type: "text_message",
		Data: "1.1 1.2 1.3 1.4 1.5 1.6 1.7",
	})
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "buy", result.Actions[0].Name)
}
