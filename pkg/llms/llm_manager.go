package llms

import (
	"context"
	"os"

	"github.com/pkg/errors"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/llms/openai"

	"github.com/yubing744/trading-gpt/pkg/config"
	"github.com/yubing744/trading-gpt/pkg/llms/anthropic"
)

type LLMManager struct {
	cfg     *config.LLMConfig
	llms    map[string]llms.Model
	primary string
}

func NewLLMManager(cfg *config.LLMConfig) *LLMManager {
	return &LLMManager{
		cfg:     cfg,
		llms:    make(map[string]llms.Model, 0),
		primary: cfg.Primary,
	}
}

func (mgr *LLMManager) Init() error {

	// init openai model
	if mgr.cfg.OpenAI != nil {
		openAICfg := mgr.cfg.OpenAI

		token := os.Getenv("LLM_OPENAI_TOKEN")
		if token == "" {
			return errors.New("AGENT_OPENAI_TOKEN not set in .env.local")
		}
		openAICfg.Token = token

		opts := make([]openai.Option, 0)
		opts = append(opts, openai.WithToken(openAICfg.Token))
		opts = append(opts, openai.WithModel(openAICfg.Model))

		if openAICfg.BaseURL != "" {
			opts = append(opts, openai.WithBaseURL(openAICfg.BaseURL))
		}

		llm, err := openai.New(opts...)
		if err != nil {
			return errors.Wrap(err, "New openai fail")
		}

		mgr.llms["openai"] = llm
	}

	// init anthropic model
	if mgr.cfg.Anthropic != nil {
		anthropicCfg := mgr.cfg.Anthropic

		token := os.Getenv("LLM_ANTHROPIC_TOKEN")
		if token == "" {
			return errors.New("LLM_ANTHROPIC_TOKEN not set in .env.local")
		}
		anthropicCfg.Token = token

		opts := make([]anthropic.Option, 0)
		opts = append(opts, anthropic.WithToken(anthropicCfg.Token))

		llm, err := anthropic.New(anthropicCfg.Model, opts...)
		if err != nil {
			return errors.Wrap(err, "New anthropic AI fail")
		}

		mgr.llms["anthropic"] = llm
	}

	// init ollama model
	if mgr.cfg.Ollama != nil {
		ollamaCfg := mgr.cfg.Ollama

		opts := make([]ollama.Option, 0)
		opts = append(opts, ollama.WithServerURL(ollamaCfg.ServerURL))
		opts = append(opts, ollama.WithModel(ollamaCfg.Model))

		llm, err := ollama.New(opts...)
		if err != nil {
			return errors.Wrap(err, "New ollama AI fail")
		}

		mgr.llms["ollama"] = llm
	}

	return nil
}

func (mgr *LLMManager) GetLLM() (llms.Model, error) {
	llm, ok := mgr.llms[mgr.primary]
	if !ok {
		return nil, errors.New("no primary llm")
	}

	return llm, nil
}

// GenerateContent asks the model to generate content from a sequence of
// messages. It's the most general interface for multi-modal LLMs that support
// chat-like interactions.
func (mgr *LLMManager) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) {
	llm, err := mgr.GetLLM()
	if err != nil {
		return nil, errors.Wrap(err, "get llm fail")
	}

	return llm.GenerateContent(ctx, messages, options...)
}

// Call is a simplified interface for a text-only Model, generating a single
// string response from a single string prompt.
//
// Deprecated: this method is retained for backwards compatibility. Use the
// more general [GenerateContent] instead. You can also use
// the [GenerateFromSinglePrompt] function which provides a similar capability
// to Call and is built on top of the new interface.
func (mgr *LLMManager) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	llm, err := mgr.GetLLM()
	if err != nil {
		return "", errors.Wrap(err, "get llm fail")
	}

	return llm.Call(ctx, prompt, options...)
}
