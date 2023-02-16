package openai

import (
	"context"
	"fmt"
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
	client           *gogpt.Client
	name             string
	backgroup        string
	chats            []string
	actions          map[string]*types.ActionDesc
	maxContextLength int
}

func NewOpenAIAgent(cfg *config.AgentOpenAIConfig) *OpenAIAgent {
	client := gogpt.NewClient(cfg.Token)

	return &OpenAIAgent{
		client:           client,
		name:             "AI",
		backgroup:        "",
		chats:            make([]string, 0),
		actions:          make(map[string]*types.ActionDesc, 0),
		maxContextLength: cfg.MaxContextLength,
	}
}

func (agent *OpenAIAgent) toPrompt(event *types.Event) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("%s:%s", HumanLable, event.Data))
	builder.WriteString("\n")

	builder.WriteString(fmt.Sprintf("%s:", agent.name))

	return builder.String()
}

func (agent *OpenAIAgent) splitChatsByLength(sessionChats []string, maxLength int) []string {
	length := 0

	for i := len(sessionChats) - 1; i >= 0; i-- {
		if length+len(sessionChats[i])+1 > maxLength {
			return sessionChats[i+1:]
		} else {
			length = length + len(sessionChats[i])
		}
	}

	return sessionChats
}

func (agent *OpenAIAgent) GenPrompt(sessionChats []string, event *types.Event) (string, error) {
	var builder strings.Builder

	builder.WriteString(agent.backgroup)
	builder.WriteString("\n\n")

	for _, chat := range agent.chats {
		builder.WriteString(chat)
		builder.WriteString("\n")
	}

	eventPrompt := agent.toPrompt(event)

	subChats := agent.splitChatsByLength(sessionChats, agent.maxContextLength-builder.Len()-len(eventPrompt))

	for _, chat := range subChats {
		builder.WriteString(chat)
		builder.WriteString("\n")
	}

	builder.WriteString(eventPrompt)

	prompt := builder.String()

	if len(prompt) > agent.maxContextLength {
		return "", errors.New("Gen prompt too long")
	}

	return prompt, nil
}

func (a *OpenAIAgent) SetName(name string) {
	a.name = name
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

func (a *OpenAIAgent) GenActions(ctx context.Context, session types.ISession, event *types.Event) (*agent.GenResult, error) {
	prompt, err := a.GenPrompt(session.GetChats(), event)
	if err != nil {
		return nil, err
	}

	req := gogpt.CompletionRequest{
		Model:            gogpt.GPT3TextDavinci003,
		Temperature:      0.5,
		MaxTokens:        256,
		TopP:             0.3,
		FrequencyPenalty: 0.5,
		PresencePenalty:  0,
		Prompt:           prompt,
	}

	resp, err := a.client.CreateCompletion(ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, "create CreateCompletionStream error")
	}

	result := &agent.GenResult{
		Actions: make([]*types.Action, 0),
		Texts:   make([]string, 0),
	}

	if len(resp.Choices) > 0 {
		text := resp.Choices[0].Text
		log.WithField("text", text).Info("resp.Choices[0].Text")

		result.Texts = append(result.Texts, text)

		for _, actionDef := range a.actions {
			if strings.Contains(text, actionDef.Name) {
				result.Actions = append(result.Actions, &types.Action{
					Name: actionDef.Name,
					Args: []string{},
				})
			}
		}
	}

	if len(result.Texts) > 0 {
		session.AddChat(fmt.Sprintf("%s:%s", HumanLable, event.Data))
		session.AddChat(fmt.Sprintf("%s:%s", a.name, strings.Join(result.Texts, "")))
	}

	return result, nil
}
