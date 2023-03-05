package openai

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/yubing744/trading-gpt/pkg/agent"
	"github.com/yubing744/trading-gpt/pkg/config"
	"github.com/yubing744/trading-gpt/pkg/types"

	gogpt "github.com/sashabaranov/go-openai"
)

const (
	HumanLable = "You"
)

var log = logrus.WithField("agent", "openai")

type OpenAIAgent struct {
	client           *gogpt.Client
	name             string
	model            string
	temperature      float32
	backgroup        string
	chats            []string
	actions          map[string]*types.ActionDesc
	maxContextLength int
}

func NewOpenAIAgent(cfg *config.AgentOpenAIConfig) *OpenAIAgent {
	client := gogpt.NewClient(cfg.Token)

	return &OpenAIAgent{
		client:           client,
		name:             cfg.Name,
		model:            cfg.Model,
		temperature:      cfg.Temperature,
		backgroup:        cfg.Backgroup,
		chats:            make([]string, 0),
		actions:          make(map[string]*types.ActionDesc, 0),
		maxContextLength: cfg.MaxContextLength,
	}
}

func (agent *OpenAIAgent) toPrompt(msgs []*types.Message) string {
	var builder strings.Builder

	for _, msg := range msgs {
		builder.WriteString(fmt.Sprintf("%s:%s", HumanLable, msg.Text))
		builder.WriteString("\n")
	}

	builder.WriteString(fmt.Sprintf("%s:", agent.name))

	return builder.String()
}

func (agent *OpenAIAgent) splitChatsByLength(sessionChats []string, maxLength int) []string {
	length := 0

	for i := len(sessionChats) - 1; i >= 0; i-- {
		if length+len(sessionChats[i])+1 > maxLength {
			return sessionChats[i+1:]
		} else {
			length = length + len(sessionChats[i]) + 1
		}
	}

	return sessionChats
}

func (agent *OpenAIAgent) GenPrompt(sessionChats []string, msgs []*types.Message) (string, error) {
	var builder strings.Builder

	// Backgougroup
	builder.WriteString(agent.backgroup)
	builder.WriteString("\n")

	// cmd help
	for _, def := range agent.actions {
		builder.WriteString(fmt.Sprintf("命令 /%s 表示：%s\n", def.Name, def.Description))
	}

	builder.WriteString("\n\n")

	// sample
	for _, chat := range agent.chats {
		builder.WriteString(chat)
		builder.WriteString("\n")
	}

	eventPrompt := agent.toPrompt(msgs)

	if builder.Len()+len(eventPrompt) > agent.maxContextLength {
		return "", errors.Errorf("Current msgs too long, current: %d, left: %d", len(eventPrompt), agent.maxContextLength-builder.Len())
	}

	subChats := agent.splitChatsByLength(sessionChats, agent.maxContextLength-builder.Len()-len(eventPrompt))

	for _, chat := range subChats {
		builder.WriteString(chat)
		builder.WriteString("\n")
	}

	builder.WriteString(eventPrompt)

	prompt := builder.String()

	if len(prompt) > agent.maxContextLength {
		return "", errors.Errorf("Gen prompt too long, current: %d, max: %d", len(prompt), agent.maxContextLength)
	}

	return prompt, nil
}

func (agent *OpenAIAgent) GenChatgptMessages(sessionChats []string, msgs []*types.Message) ([]gogpt.ChatCompletionMessage, error) {
	chatgptMsgs := make([]gogpt.ChatCompletionMessage, 0)

	// Backgougroup
	chatgptMsgs = append(chatgptMsgs, gogpt.ChatCompletionMessage{
		Role:    "system",
		Content: agent.backgroup,
	})

	for _, msg := range msgs {
		chatgptMsgs = append(chatgptMsgs, gogpt.ChatCompletionMessage{
			Role:    "user",
			Content: msg.Text,
		})
	}

	return chatgptMsgs, nil
}

func (a *OpenAIAgent) GetName() string {
	return fmt.Sprintf("openai-%s", a.model)
}

func (a *OpenAIAgent) SetBackgroup(backgroup string) {
	a.backgroup = backgroup
}

func (a *OpenAIAgent) RegisterActions(ctx context.Context, name string, actions []*types.ActionDesc) {
	for _, def := range actions {
		a.actions[def.Name] = def

		for _, sample := range def.Samples {
			for _, input := range sample.Input {
				a.chats = append(a.chats, fmt.Sprintf("%s:%s", HumanLable, input))
			}

			for _, output := range sample.Output {
				a.chats = append(a.chats, fmt.Sprintf("%s:%s", a.name, output))
			}
		}
	}
}

func (a *OpenAIAgent) Start() error {
	return nil
}

func (a *OpenAIAgent) Stop() {

}

func (a *OpenAIAgent) GenActions(ctx context.Context, session types.ISession, msgs []*types.Message) (*agent.GenResult, error) {
	gptMsgs, err := a.GenChatgptMessages(session.GetChats(), msgs)
	if err != nil {
		return nil, err
	}

	log.
		WithField("chatgpt msgs", gptMsgs).
		Infof("gen chatgpt messages")

	req := gogpt.ChatCompletionRequest{
		Model:       a.model,
		Temperature: a.temperature,
		Messages:    gptMsgs,
	}

	resp, err := a.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, "create CreateCompletionStream error")
	}

	result := &agent.GenResult{
		Texts: make([]string, 0),
	}

	if len(resp.Choices) > 0 {
		text := resp.Choices[0].Message.Content
		log.WithField("text", text).Info("resp.Choices[0].Text")

		result.Texts = append(result.Texts, text)
	}

	if len(result.Texts) > 0 {
		for _, msg := range msgs {
			session.AddChat(fmt.Sprintf("%s:%s", HumanLable, msg.Text))
		}

		session.AddChat(fmt.Sprintf("%s:%s", a.name, strings.Join(result.Texts, "")))
	}

	return result, nil
}
