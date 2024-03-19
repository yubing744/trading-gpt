package anthropic

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/madebywelch/anthropic-go/v2/pkg/anthropic"
	"github.com/sirupsen/logrus"
	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
)

var (
	ErrEmptyResponse = errors.New("no response")
	ErrMissingToken  = errors.New("missing the Anthropic API key, set it in the ANTHROPIC_API_KEY environment variable")

	ErrUnexpectedResponseLength = errors.New("unexpected length of response")
)

var log = logrus.WithField("llm", "anthropic")

type LLM struct {
	CallbacksHandler callbacks.Handler
	client           *anthropic.Client
	model            string
}

var _ llms.Model = (*LLM)(nil)

// New returns a new Anthropic LLM.
func New(model string, opts ...Option) (*LLM, error) {
	c, err := newClient(opts...)

	return &LLM{
		model:  model,
		client: c,
	}, err
}

func newClient(opts ...Option) (*anthropic.Client, error) {
	options := &options{
		token: os.Getenv(tokenEnvVarName),
	}

	for _, opt := range opts {
		opt(options)
	}

	if len(options.token) == 0 {
		return nil, ErrMissingToken
	}

	log.WithField("token", options.token).Info("anthropic_new")

	return anthropic.NewClient(options.token)
}

// Call requests a completion for the given prompt.
func (o *LLM) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	return llms.GenerateFromSinglePrompt(ctx, o, prompt, options...)
}

// GenerateContent implements the Model interface.
func (o *LLM) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) { //nolint: lll, cyclop, whitespace

	if o.CallbacksHandler != nil {
		o.CallbacksHandler.HandleLLMGenerateContentStart(ctx, messages)
	}

	opts := &llms.CallOptions{}
	for _, opt := range options {
		opt(opts)
	}

	systemPrompt := ""

	// Assume we get a single text message
	chatMsgs := make([]anthropic.MessagePartRequest, 0)

	for _, mc := range messages {
		textMsg := joinTextParts(mc.Parts)
		msg := anthropic.MessagePartRequest{
			Content: textMsg,
		}

		switch mc.Role {
		case schema.ChatMessageTypeSystem:
			systemPrompt = textMsg
			continue
		case schema.ChatMessageTypeAI:
			msg.Role = "assistant"
		case schema.ChatMessageTypeHuman:
			msg.Role = "user"
		default:
			return nil, fmt.Errorf("role %v not supported", mc.Role)
		}

		chatMsgs = append(chatMsgs, msg)
	}

	anthropicOpts := make([]anthropic.GenericOption[anthropic.MessageRequest], 0)
	anthropicOpts = append(anthropicOpts, anthropic.WithMaxTokens[anthropic.MessageRequest](opts.MaxTokens))

	if opts.Model == "" {
		opts.Model = o.model
	}

	anthropicOpts = append(anthropicOpts, anthropic.WithModel[anthropic.MessageRequest](anthropic.Model(opts.Model)))

	if systemPrompt != "" {
		anthropicOpts = append(anthropicOpts, anthropic.WithSystemPrompt(systemPrompt))
	}

	// Call the Message method
	req := anthropic.NewMessageRequest(
		chatMsgs,
		anthropicOpts...,
	)

	log.WithField("req", req).Info("GenerateContent_req")

	response, err := o.client.Message(req)
	if err != nil {
		if o.CallbacksHandler != nil {
			o.CallbacksHandler.HandleLLMError(ctx, err)
		}
		return nil, err
	}

	content := ""
	for _, part := range response.Content {
		content = content + part.Text
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
		textPart, ok := part.(llms.TextContent)
		if ok {
			text += textPart.Text
		}
	}

	return text
}
