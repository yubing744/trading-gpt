package openai

import (
	"context"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

type LLM struct {
	innerLLM     *openai.LLM
	noSystemRole bool
}

func New(noSystemRole bool, opts ...openai.Option) (*LLM, error) {
	llm, err := openai.New(opts...)
	if err != nil {
		return nil, err
	}

	return &LLM{
		innerLLM:     llm,
		noSystemRole: noSystemRole,
	}, nil
}

func (llm *LLM) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) {
	if llm.noSystemRole {
		msgs := make([]llms.MessageContent, 0)

		for _, msg := range messages {
			if msg.Role == llms.ChatMessageTypeSystem {
				msgs = append(msgs, llms.MessageContent{
					Role:  llms.ChatMessageTypeHuman,
					Parts: msg.Parts,
				})
			} else {
				msgs = append(msgs, msg)
			}
		}

		return llm.innerLLM.GenerateContent(ctx, msgs, options...)
	}

	return llm.innerLLM.GenerateContent(ctx, messages, options...)
}

func (llm *LLM) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	return llm.innerLLM.Call(ctx, prompt, options...)
}
