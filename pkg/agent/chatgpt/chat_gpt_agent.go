package chatgpt

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/yubing744/trading-bot/pkg/agent"
	"github.com/yubing744/trading-bot/pkg/config"
	"github.com/yubing744/trading-bot/pkg/types"

	gpt "github.com/yubing744/chatgpt-go/pkg"
	gptcfg "github.com/yubing744/chatgpt-go/pkg/config"
)

const (
	HumanLable = "You"
)

var log = logrus.WithField("agent", "chatgpt")

type ChatGPTAgent struct {
	client *gpt.ChatgptClient

	name             string
	backgroup        string
	chats            []string
	actions          map[string]*types.ActionDesc
	maxContextLength int

	conversations map[string]*ChatGPTConversation
	lock          sync.RWMutex
}

func NewChatGPTAgent(cfg *config.AgentChatGPTConfig) *ChatGPTAgent {
	chatgptCfg := &gptcfg.Config{
		Email:    cfg.Email,
		Password: cfg.Password,
		Proxy:    "",
		Timeout:  time.Second * 600,
		Debug:    false,
	}

	client := gpt.NewChatgptClient(chatgptCfg)

	return &ChatGPTAgent{
		client:           client,
		name:             "AI",
		backgroup:        "",
		chats:            make([]string, 0),
		actions:          make(map[string]*types.ActionDesc, 0),
		maxContextLength: cfg.MaxContextLength,
		conversations:    make(map[string]*ChatGPTConversation),
	}
}

func (a *ChatGPTAgent) SetName(name string) {
	a.name = name
}

func (a *ChatGPTAgent) SetBackgroup(backgroup string) {
	a.backgroup = backgroup
}

func (a *ChatGPTAgent) RegisterActions(ctx context.Context, name string, actions []*types.ActionDesc) {
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

func (agent *ChatGPTAgent) toPrompt(msgs []*types.Message) string {
	var builder strings.Builder

	for _, msg := range msgs {
		builder.WriteString(fmt.Sprintf("%s:%s", HumanLable, msg.Text))
		builder.WriteString("\n")
	}

	builder.WriteString(fmt.Sprintf("%s:", agent.name))

	return builder.String()
}

func (agent *ChatGPTAgent) genInitPrompt(conv *ChatGPTConversation, msgs []*types.Message) (string, error) {
	var builder strings.Builder

	if conv.IsNew() {
		// Backgougroup
		builder.WriteString(agent.backgroup)
		builder.WriteString("\n")

		// cmd help
		for _, def := range agent.actions {
			builder.WriteString(fmt.Sprintf("The trading cmd /%s means: %s\n", def.Name, def.Description))
		}

		builder.WriteString("\n\n")

		// sample
		for _, chat := range agent.chats {
			builder.WriteString(chat)
			builder.WriteString("\n")
		}
	}

	builder.WriteString(agent.toPrompt(msgs))

	prompt := builder.String()

	if len(prompt) > agent.maxContextLength {
		return "", errors.Errorf("Gen prompt too long, current: %d, max: %d", len(prompt), agent.maxContextLength)
	}

	return prompt, nil
}

func (a *ChatGPTAgent) Init() error {
	log.Info("Login ...")
	err := a.client.Login(context.Background())
	if err != nil {
		return err
	}

	log.Info("Login success.")

	return nil
}

func (agent *ChatGPTAgent) getOrCreate(sessionId string) *ChatGPTConversation {
	agent.lock.Lock()
	defer agent.lock.Unlock()

	conv, ok := agent.conversations[sessionId]
	if !ok {
		conv = NewChatGPTConversation()
		agent.conversations[sessionId] = conv
	}

	return conv
}

func (a *ChatGPTAgent) GenActions(ctx context.Context, session types.ISession, msgs []*types.Message) (*agent.GenResult, error) {
	conv := a.getOrCreate(session.GetID())

	prompt, err := a.genInitPrompt(conv, msgs)
	if err != nil {
		return nil, err
	}

	log.
		WithField("conv_id", conv.GetID()).
		WithField("parent_messsage_id", conv.GetParentMessageID()).
		WithField("prompt_length", len(prompt)).
		WithField("max_length", a.maxContextLength).
		Infof("gen prompt")

	log.Info(prompt)

	resp, err := a.client.Ask(context.Background(), prompt, conv.GetIDRef(), conv.GetParentMessageIDRef(), time.Second*20)
	if err != nil {
		return nil, err
	}

	result := &agent.GenResult{
		Actions: make([]*types.Action, 0),
		Texts:   make([]string, 0),
	}

	if resp.Code == 0 {
		conv.Update(resp.Data.ConversationID, resp.Data.ParentID)

		text := resp.Data.Text
		log.WithField("text", text).Info("resp.Data.Text")

		result.Texts = append(result.Texts, text)

		for _, actionDef := range a.actions {
			if strings.Contains(strings.ToLower(text), actionDef.Name) {
				log.WithField("action", actionDef.Name).Info("match action")

				result.Actions = append(result.Actions, &types.Action{
					Target: "exchange",
					Name:   actionDef.Name,
					Args:   []string{},
				})
			}
		}
	}

	return result, nil
}
