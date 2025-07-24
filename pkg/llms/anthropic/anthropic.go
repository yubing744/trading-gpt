package anthropic

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/sirupsen/logrus"
	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/llms"
)

const DefaultTokenSample = 4000

var (
	ErrEmptyResponse = errors.New("no response")
	ErrMissingToken  = errors.New("missing the Anthropic API key, set it in the ANTHROPIC_API_KEY environment variable")
)

var log = logrus.WithField("llm", "anthropic")

type LLM struct {
	CallbacksHandler callbacks.Handler
	client           *anthropic.Client
	model            string
	thinkingBudget   int64
	enableThinking   bool
}

var _ llms.Model = (*LLM)(nil)

// New returns a new Anthropic LLM.
func New(model string, opts ...Option) (*LLM, error) {
	c, err := newClient(opts...)
	if err != nil {
		return nil, err
	}

	options := &options{}
	for _, opt := range opts {
		opt(options)
	}

	return &LLM{
		model:          model,
		client:         c,
		thinkingBudget: options.thinkingBudget,
		enableThinking: options.enableThinking,
	}, nil
}

func newClient(opts ...Option) (*anthropic.Client, error) {
	options := &options{
		baseURL: "https://api.anthropic.com",
		token:   os.Getenv(tokenEnvVarName),
	}

	for _, opt := range opts {
		opt(options)
	}

	if len(options.token) == 0 {
		return nil, ErrMissingToken
	}

	log.WithField("token", options.token[:8]+"...").Info("anthropic_new")

	client := anthropic.NewClient(
		option.WithAPIKey(options.token),
		option.WithBaseURL(options.baseURL),
	)

	return &client, nil
}

// Call requests a completion for the given prompt.
func (o *LLM) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	return llms.GenerateFromSinglePrompt(ctx, o, prompt, options...)
}

// GenerateContent implements the Model interface.
func (o *LLM) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) {
	if o.CallbacksHandler != nil {
		o.CallbacksHandler.HandleLLMGenerateContentStart(ctx, messages)
	}

	opts := &llms.CallOptions{}
	for _, opt := range options {
		opt(opts)
	}

	// Build system prompt and messages
	systemPrompt := ""
	var anthropicMessages []anthropic.MessageParam

	for _, mc := range messages {
		textMsg := joinTextParts(mc.Parts)

		switch mc.Role {
		case llms.ChatMessageTypeSystem:
			systemPrompt = textMsg
		case llms.ChatMessageTypeAI:
			anthropicMessages = append(anthropicMessages, anthropic.NewAssistantMessage(anthropic.NewTextBlock(textMsg)))
		case llms.ChatMessageTypeHuman:
			anthropicMessages = append(anthropicMessages, anthropic.NewUserMessage(anthropic.NewTextBlock(textMsg)))
		default:
			return nil, fmt.Errorf("role %v not supported", mc.Role)
		}
	}

	// Prepare request parameters
	maxTokens := opts.MaxTokens
	if maxTokens <= 0 {
		maxTokens = DefaultTokenSample
	}

	if opts.Model == "" {
		opts.Model = o.model
	}

	// Create message request
	req := anthropic.MessageNewParams{
		MaxTokens: int64(maxTokens),
		Messages:  anthropicMessages,
		Model:     anthropic.Model(opts.Model),
	}

	if systemPrompt != "" {
		req.System = []anthropic.TextBlockParam{
			{Text: systemPrompt},
		}
	}

	// Add extended thinking configuration
	if o.enableThinking && o.thinkingBudget > 0 {
		req.Thinking = anthropic.ThinkingConfigParamOfEnabled(o.thinkingBudget)
	}

	// Call the API
	response, err := o.client.Messages.New(ctx, req)
	if err != nil {
		if o.CallbacksHandler != nil {
			o.CallbacksHandler.HandleLLMError(ctx, err)
		}
		return nil, err
	}

	// Extract content from response
	content := ""
	if len(response.Content) > 0 {
		for _, c := range response.Content {
			if c.Type == "thinking" {
				content += "<thinking>" + c.Thinking + "</thinking>\n"
			} else if c.Type == "text" {
				content += c.Text
			}
		}
	}

	resp := &llms.ContentResponse{
		Choices: []*llms.ContentChoice{
			{
				Content: content,
			},
		},
	}

	return resp, nil
}

func joinTextParts(parts []llms.ContentPart) string {
	text := ""
	for _, part := range parts {
		if textPart, ok := part.(llms.TextContent); ok {
			text += textPart.Text
		}
	}
	return text
}
