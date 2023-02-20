package chatgpt

import (
	"context"
	"fmt"
	"strings"
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
}

func NewChatGPTAgent(cfg *config.AgentChatGPTConfig) *ChatGPTAgent {
	chatgptCfg := &gptcfg.Config{
		Email:    cfg.Email,
		Password: cfg.Password,
	}

	client := gpt.NewChatgptClient(chatgptCfg)

	return &ChatGPTAgent{
		client:           client,
		name:             "AI",
		backgroup:        "",
		chats:            make([]string, 0),
		actions:          make(map[string]*types.ActionDesc, 0),
		maxContextLength: cfg.MaxContextLength,
	}
}

func (agent *ChatGPTAgent) SetName(name string) {
	agent.name = name
}

func (agent *ChatGPTAgent) SetBackgroup(backgroup string) {
	agent.backgroup = backgroup
}

func (agent *ChatGPTAgent) RegisterActions(ctx context.Context, name string, actions []*types.ActionDesc) {

}

func (agent *ChatGPTAgent) genInitPrompt() (string, error) {
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

	builder.WriteString(fmt.Sprintf("%s:%s", HumanLable, "Hello"))
	builder.WriteString(fmt.Sprintf("%s:", agent.name))

	prompt := builder.String()

	if len(prompt) > agent.maxContextLength {
		return "", errors.Errorf("Gen prompt too long, current: %d, max: %d", len(prompt), agent.maxContextLength)
	}

	return prompt, nil
}

func (a *ChatGPTAgent) Init() error {
	err := a.client.Login()
	if err != nil {
		return err
	}

	prompt, err := a.genInitPrompt()
	if err != nil {
		return err
	}

	result, err := a.client.Ask(context.Background(), prompt, nil, nil, time.Second*20)
	if err != nil {
		return err
	}

	if result.Code == 0 {
		log.Infof("%s:%s", a.name, result.Data.Text)
	}

	return nil
}

func (agent *ChatGPTAgent) GenActions(ctx context.Context, session types.ISession, msgs []*types.Message) (*agent.GenResult, error) {
	return nil, nil
}
