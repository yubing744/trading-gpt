package anthropic

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/madebywelch/anthropic-go/v3/pkg/anthropic"
	"github.com/madebywelch/anthropic-go/v3/pkg/anthropic/client/native"
	"github.com/sirupsen/logrus"
	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/llms"
)

const MaxTokenSample = 4000

var (
	ErrEmptyResponse = errors.New("no response")
	ErrMissingToken  = errors.New("missing the Anthropic API key, set it in the ANTHROPIC_API_KEY environment variable")

	ErrUnexpectedResponseLength = errors.New("unexpected length of response")
)

var log = logrus.WithField("llm", "anthropic")

type LLM struct {
	CallbacksHandler callbacks.Handler
	client           *native.Client
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

func newClient(opts ...Option) (*native.Client, error) {
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

	log.WithField("token", options.token).Info("anthropic_new")

	return native.MakeClient(native.Config{
		BaseURL: options.baseURL,
		APIKey:  options.token,
	})
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

	chatMsgs := make([]anthropic.MessagePartRequest, 0)
	var lastRole llms.ChatMessageType
	var tmpContentBlocks []anthropic.ContentBlock

	for _, mc := range messages {
		textMsg := joinTextParts(mc.Parts)

		if mc.Role == lastRole && mc.Role == llms.ChatMessageTypeHuman {
			tmpContentBlocks = append(tmpContentBlocks, anthropic.NewTextContentBlock(textMsg))
			continue
		}

		if len(tmpContentBlocks) > 0 {
			// Append the buffered message before starting a new one
			chatMsgs = append(chatMsgs, anthropic.MessagePartRequest{
				Content: tmpContentBlocks,
				Role:    "user",
			})
			tmpContentBlocks = make([]anthropic.ContentBlock, 0)
		}

		switch mc.Role {
		case llms.ChatMessageTypeSystem:
			systemPrompt = textMsg
			continue
		case llms.ChatMessageTypeAI:
			msg := anthropic.MessagePartRequest{
				Content: []anthropic.ContentBlock{anthropic.NewTextContentBlock(textMsg)},
				Role:    "assistant",
			}
			chatMsgs = append(chatMsgs, msg)
		case llms.ChatMessageTypeHuman:
			tmpContentBlocks = append(tmpContentBlocks, anthropic.NewTextContentBlock(textMsg))
		default:
			return nil, fmt.Errorf("role %v not supported", mc.Role)
		}

		lastRole = mc.Role
	}

	if len(tmpContentBlocks) > 0 {
		chatMsgs = append(chatMsgs, anthropic.MessagePartRequest{
			Content: tmpContentBlocks,
			Role:    "user",
		})
	}

	anthropicOpts := make([]anthropic.GenericOption[anthropic.MessageRequest], 0)

	maxTokenSample := opts.MaxTokens
	if maxTokenSample > MaxTokenSample {
		maxTokenSample = MaxTokenSample
	}

	anthropicOpts = append(anthropicOpts, anthropic.WithMaxTokens[anthropic.MessageRequest](opts.MaxTokens))

	if opts.Model == "" {
		opts.Model = o.model
	}

	anthropicOpts = append(anthropicOpts, anthropic.WithModel[anthropic.MessageRequest](anthropic.Model(opts.Model)))
	anthropicOpts = append(anthropicOpts, anthropic.WithTemperature[anthropic.MessageRequest](opts.Temperature))

	if systemPrompt != "" {
		anthropicOpts = append(anthropicOpts, anthropic.WithSystemPrompt(systemPrompt))
	}

	// Call the Message method
	req := anthropic.NewMessageRequest(
		chatMsgs,
		anthropicOpts...,
	)

	response, err := o.client.Message(ctx, req)
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
