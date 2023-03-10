package openai

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yubing744/trading-gpt/pkg/config"
	"github.com/yubing744/trading-gpt/pkg/types"
)

func TestNewOpenAIAgent(t *testing.T) {
	cfg := &config.AgentOpenAIConfig{
		Name:             "AI",
		Token:            "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",
		MaxContextLength: 500,
	}
	agent := NewOpenAIAgent(cfg)
	assert.NotNil(t, agent)
}

func TestGenPrompt(t *testing.T) {
	cfg := &config.AgentOpenAIConfig{
		Name:             "AI",
		Token:            "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",
		MaxContextLength: 500,
	}
	agent := NewOpenAIAgent(cfg)
	assert.NotNil(t, agent)

	prompt, err := agent.GenPrompt([]string{}, []*types.Message{
		{
			Text: "xxx",
		},
	})

	assert.NoError(t, err)
	assert.Equal(t, "\n\n\nYou:xxx\nAI:", prompt)
}

func TestGenPromptWithTooLong(t *testing.T) {
	cfg := &config.AgentOpenAIConfig{
		Name:             "AI",
		Token:            "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",
		MaxContextLength: 13,
	}
	agent := NewOpenAIAgent(cfg)
	assert.NotNil(t, agent)

	_, err := agent.GenPrompt([]string{}, []*types.Message{
		{
			Text: "xxx",
		},
	})

	assert.Error(t, err)
}

func TestGenPromptBySplitChats(t *testing.T) {
	cfg := &config.AgentOpenAIConfig{
		Name:             "AI",
		Token:            "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",
		MaxContextLength: 14,
	}
	agent := NewOpenAIAgent(cfg)
	assert.NotNil(t, agent)

	prompt, err := agent.GenPrompt([]string{"msg xxxx"}, []*types.Message{
		{
			Text: "xxx",
		},
	})

	assert.NoError(t, err)
	assert.Equal(t, "\n\n\nYou:xxx\nAI:", prompt)
}

func TestGenPromptBySplitChats2(t *testing.T) {
	cfg := &config.AgentOpenAIConfig{
		Name:             "AI",
		Token:            "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",
		MaxContextLength: 25,
	}
	agent := NewOpenAIAgent(cfg)
	assert.NotNil(t, agent)

	prompt, err := agent.GenPrompt([]string{"msg xxxx1", "msg xxxx2"}, []*types.Message{
		{
			Text: "xxx",
		},
	})

	assert.NoError(t, err)
	assert.Equal(t, "\n\n\nmsg xxxx2\nYou:xxx\nAI:", prompt)
}

func TestGenPromptBySplitChats3(t *testing.T) {
	cfg := &config.AgentOpenAIConfig{
		Name:             "AI",
		Token:            "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",
		MaxContextLength: 100,
	}
	agent := NewOpenAIAgent(cfg)
	assert.NotNil(t, agent)

	prompt, err := agent.GenPrompt([]string{"msg xxxx1", "msg xxxx2"}, []*types.Message{
		{
			Text: "xxx",
		},
	})

	assert.NoError(t, err)
	assert.Equal(t, "\n\n\nmsg xxxx1\nmsg xxxx2\nYou:xxx\nAI:", prompt)
}

func TestGenPromptBySplitChats4(t *testing.T) {
	cfg := &config.AgentOpenAIConfig{
		Name:             "AI",
		Token:            "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",
		MaxContextLength: 23,
	}
	agent := NewOpenAIAgent(cfg)
	assert.NotNil(t, agent)

	agent.SetBackgroup("backgroup")
	prompt, err := agent.GenPrompt([]string{"msg xxxx1", "msg xxxx2"}, []*types.Message{
		{
			Text: "xxx",
		},
	})

	assert.NoError(t, err)
	assert.Equal(t, "backgroup\n\n\nYou:xxx\nAI:", prompt)
}

func TestGenPromptBySplitChats5(t *testing.T) {
	cfg := &config.AgentOpenAIConfig{
		Name:             "AI",
		Token:            "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",
		MaxContextLength: 100,
	}
	agent := NewOpenAIAgent(cfg)
	assert.NotNil(t, agent)

	agent.SetBackgroup("backgroup")

	sessionChats := make([]string, 0)
	for i := 1; i < 4000; i++ {
		sessionChats = append(sessionChats, fmt.Sprintf("test message %d", i))

		_, err := agent.GenPrompt(sessionChats, []*types.Message{
			{
				Text: "xxx",
			},
		})

		assert.NoError(t, err)
	}
}

func TestGenAction(t *testing.T) {
	ctx := context.Background()

	cfg := &config.AgentOpenAIConfig{
		Name:             "AI",
		Model:            "text-davinci-003",
		Token:            "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",
		MaxContextLength: 600,
	}
	agent := NewOpenAIAgent(cfg)
	assert.NotNil(t, agent)

	agent.SetBackgroup("????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????K????????????")
	agent.RegisterActions(ctx, "exchange", []*types.ActionDesc{
		{
			Name:        "buy",
			Description: "????????????",
			Samples: []types.Sample{
				{
					Input: []string{
						"1.0 1.1 1.2 1.3 1.4 1.5 1.6",
						"1.0 1.1 1.0 1.1 1.2 1.3 1.4",
					},
					Output: []string{
						"/buy [] #?????????????????????",
					},
				},
			},
		},
		{
			Name:        "sell",
			Description: "????????????",
			Samples: []types.Sample{
				{
					Input: []string{
						"1.6 1.5 1.4 1.3 1.2 1.1 1.0",
						"1.0 1.1 1.2 1.3 1.4 1.3 1.2",
					},
					Output: []string{
						"/buy [] #?????????????????????",
					},
				},
			},
		},
		{
			Name:        "hold",
			Description: "??????",
			Samples: []types.Sample{
				{
					Input: []string{
						"1.2 1.3 1.4 1.5 1.6 1.7 1.8",
					},
					Output: []string{
						"/buy [] #?????????????????????",
					},
				},
			},
		},
	})

	session := types.NewMockSession("session_1")

	result, err := agent.GenActions(ctx, session, []*types.Message{
		{
			Text: "1.1 1.2 1.3 1.4 1.5 1.6 1.7",
		},
	})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.GreaterOrEqual(t, len(result.Texts), 0)
}
