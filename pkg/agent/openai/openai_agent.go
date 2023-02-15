package openai

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/yubing744/trading-bot/pkg/agent"
	"github.com/yubing744/trading-bot/pkg/config"
	"github.com/yubing744/trading-bot/pkg/types"

	gogpt "github.com/sashabaranov/go-gpt3"
)

const (
	HumanLable = "You"
)

var log = logrus.WithField("agent", "openai")

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

func (a *OpenAIAgent) SetBackgroup(backgroup string) {
	a.backgroup = backgroup
}

func (a *OpenAIAgent) RegisterActions(ctx context.Context, actions []*types.ActionDesc) {
	for _, def := range actions {
		a.actions[def.Name] = def

		for _, sample := range def.Samples {
			a.chats = append(a.chats, fmt.Sprintf("%s:%s", HumanLable, sample))
			a.chats = append(a.chats, fmt.Sprintf("%s:%s", a.name, def.Name))
		}
	}
}

func (a *OpenAIAgent) GenActions(ctx context.Context, sessionID string, event *types.Event) (*agent.GenResult, error) {
	prompt := a.genPrompt(event)

	req := gogpt.CompletionRequest{
		Model:            gogpt.GPT3TextDavinci003,
		Temperature:      0.5,
		MaxTokens:        256,
		TopP:             0.3,
		FrequencyPenalty: 0.5,
		PresencePenalty:  0,
		Prompt:           prompt,
		Stream:           true,
	}

	stream, err := a.client.CreateCompletionStream(ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, "create CreateCompletionStream error")
	}

	defer stream.Close()

	result := &agent.GenResult{
		Actions: make([]*types.Action, 0),
		Texts:   make([]string, 0),
	}

	for {
		resp, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return nil, errors.Wrap(err, "read stream error")
		}

		if len(resp.Choices) > 0 {
			text := resp.Choices[0].Text
			log.WithField("text", text).Info("resp.Choices[0].Text")

			result.Texts = append(result.Texts, text)

			actionDef, ok := a.actions[text]
			if ok {
				result.Actions = append(result.Actions, &types.Action{
					Name: actionDef.Name,
					Args: []string{},
				})
			}
		}
	}

	if len(result.Texts) > 0 {
		a.chats = append(a.chats, fmt.Sprintf("%s:%s", HumanLable, event.Data))
		a.chats = append(a.chats, fmt.Sprintf("%s:%s", a.name, strings.Join(result.Texts, "")))
	}

	return result, nil
}
