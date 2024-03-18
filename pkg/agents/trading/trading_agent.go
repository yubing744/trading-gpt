package trading

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"

	"github.com/yubing744/trading-gpt/pkg/agents"
	"github.com/yubing744/trading-gpt/pkg/config"
	"github.com/yubing744/trading-gpt/pkg/types"
)

const (
	HumanLable = "You"
)

var log = logrus.WithField("agent", "openai")

type TradingAgent struct {
	llm              llms.Model
	name             string
	model            string
	temperature      float32
	maxContextLength int
	backgroup        string
	chats            []string
	actions          map[string]*types.ActionDesc
}

func NewTradingAgent(cfg *config.TradingAgentConfig, llm llms.Model) *TradingAgent {
	return &TradingAgent{
		llm:              llm,
		name:             cfg.Name,
		model:            cfg.Model,
		temperature:      cfg.Temperature,
		backgroup:        cfg.Backgroup,
		chats:            make([]string, 0),
		actions:          make(map[string]*types.ActionDesc, 0),
		maxContextLength: cfg.MaxContextLength,
	}
}

func (agent *TradingAgent) toPrompt(msgs []*types.Message) string {
	var builder strings.Builder

	for _, msg := range msgs {
		builder.WriteString(fmt.Sprintf("%s:%s", HumanLable, msg.Text))
		builder.WriteString("\n")
	}

	builder.WriteString(fmt.Sprintf("%s:", agent.name))

	return builder.String()
}

func (agent *TradingAgent) splitChatsByLength(sessionChats []string, maxLength int) []string {
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

func (agent *TradingAgent) GenPrompt(sessionChats []string, msgs []*types.Message) (string, error) {
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

func (a *TradingAgent) GetName() string {
	return fmt.Sprintf("openai-%s", a.model)
}

func (a *TradingAgent) SetBackgroup(backgroup string) {
	a.backgroup = backgroup
}

func (a *TradingAgent) RegisterActions(ctx context.Context, name string, actions []*types.ActionDesc) {
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

func (a *TradingAgent) Start() error {
	return nil
}

func (a *TradingAgent) Stop() {

}

func (a *TradingAgent) GenActions(ctx context.Context, session types.ISession, msgs []*types.Message) (*agents.GenResult, error) {
	gptMsgs, err := a.GenLLMMessages(session.GetChats(), msgs)
	if err != nil {
		return nil, err
	}

	log.
		WithField("chatgpt msgs", gptMsgs).
		Infof("gen chatgpt messages")

	callOpts := make([]llms.CallOption, 0)
	callOpts = append(callOpts, llms.WithTemperature(float64(a.temperature)))
	callOpts = append(callOpts, llms.WithMaxTokens(a.maxContextLength))

	if a.model != "" {
		callOpts = append(callOpts, llms.WithModel(a.model))
	}

	resp, err := a.llm.GenerateContent(ctx, gptMsgs, callOpts...)
	if err != nil {
		return nil, errors.Wrap(err, "create CreateCompletionStream error")
	}

	result := &agents.GenResult{
		Texts: make([]string, 0),
	}

	if len(resp.Choices) > 0 {
		text := resp.Choices[0].Content
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

func (agent *TradingAgent) GenLLMMessages(sessionChats []string, msgs []*types.Message) ([]llms.MessageContent, error) {
	llmMsgs := make([]llms.MessageContent, 0)

	// Backgougroup
	llmMsgs = append(llmMsgs, llms.MessageContent{
		Role: schema.ChatMessageTypeSystem,
		Parts: []llms.ContentPart{
			llms.TextContent{
				Text: agent.backgroup,
			},
		},
	})

	for _, msg := range msgs {
		llmMsgs = append(llmMsgs, llms.MessageContent{
			Role: schema.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextContent{
					Text: msg.Text,
				},
			},
		})
	}

	return llmMsgs, nil
}
