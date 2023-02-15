package openai

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/pkg/errors"
	"github.com/yubing744/trading-bot/pkg/config"
	"github.com/yubing744/trading-bot/pkg/types"

	gogpt "github.com/sashabaranov/go-gpt3"
)

const (
	HumanLable = "You"
)

type OpenAIAgent struct {
	client    *gogpt.Client
	name      string
	backgroup string
	chats     []string
	actions   map[string]*types.ActionDesc
}

func NewOpenAIAgent(cfg *config.AgentOpenAIConfig) *OpenAIAgent {
	client := gogpt.NewClient(cfg.Token)

	return &OpenAIAgent{
		client:    client,
		name:      "AI",
		backgroup: "",
		chats:     make([]string, 0),
		actions:   make(map[string]*types.ActionDesc, 0),
	}
}

func (agent *OpenAIAgent) genPrompt(event *types.Event) string {
	var builder strings.Builder

	builder.WriteString(agent.backgroup)
	builder.WriteString("\n\n")

	for _, chat := range agent.chats {
		builder.WriteString(chat)
		builder.WriteString("\n")
	}

	builder.WriteString(fmt.Sprintf("%s:%s", HumanLable, event.Data))
	builder.WriteString("\n")

	builder.WriteString(fmt.Sprintf("%s:", agent.name))

	return builder.String()
}

func (agent *OpenAIAgent) SetBackgroup(backgroup string) {
	agent.backgroup = backgroup
}

func (agent *OpenAIAgent) RegisterActions(ctx context.Context, actions []*types.ActionDesc) {
	for _, def := range actions {
		agent.actions[def.Name] = def

		for _, sample := range def.Samples {
			agent.chats = append(agent.chats, fmt.Sprintf("%s:%s", HumanLable, sample))
			agent.chats = append(agent.chats, fmt.Sprintf("%s:%s", agent.name, def.Name))
		}
	}
}

func (agent *OpenAIAgent) GenActions(ctx context.Context, sessionID string, event *types.Event) ([]*types.Action, error) {
	prompt := agent.genPrompt(event)

	req := gogpt.CompletionRequest{
		Model:            gogpt.GPT3TextDavinci003,
		Temperature:      0.5,
		MaxTokens:        5,
		TopP:             0.3,
		FrequencyPenalty: 0.5,
		PresencePenalty:  0,
		Prompt:           prompt,
		Stream:           true,
	}

	stream, err := agent.client.CreateCompletionStream(ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, "create CreateCompletionStream error")
	}

	defer stream.Close()

	result := make([]*types.Action, 0)

	for {
		resp, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return nil, errors.Wrap(err, "read stream error")
		}

		if len(resp.Choices) > 0 {
			actionDef, ok := agent.actions[resp.Choices[0].Text]
			if ok {
				result = append(result, &types.Action{
					Name: actionDef.Name,
					Args: []string{},
				})
			}
		}
	}

	return result, nil
}
