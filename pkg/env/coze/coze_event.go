package coze

import (
	"fmt"
	"strings"

	"github.com/yubing744/trading-gpt/pkg/apis/coze"
	"github.com/yubing744/trading-gpt/pkg/config"
	"github.com/yubing744/trading-gpt/pkg/types"
)

// CozeEvent represents an event that is specific to interactions with the Coze platform.
type CozeEvent struct {
	types.Event // Embed the base Event struct to reuse its implementation.
	config      *config.IndicatorItem
	chatResp    *coze.ChatResponse
}

// NewCozeEvent creates a new instance of CozeEvent with the given type and data.
func NewCozeEvent(config *config.IndicatorItem, chatResp *coze.ChatResponse) *CozeEvent {
	return &CozeEvent{
		Event:    *types.NewEvent(config.Name, chatResp),
		config:   config,
		chatResp: chatResp,
	}
}

// ToPrompts is overridden to include the message content in the prompts for CozeEvent.
func (e *CozeEvent) ToPrompts() []string {
	sb := strings.Builder{}

	sb.WriteString(fmt.Sprintf("%s\n", e.config.Description))

	for _, msg := range e.chatResp.Messages {
		if msg.Role == "assistant" && msg.Type == "answer" {
			sb.WriteString(msg.Content)
		}
	}

	return []string{sb.String()}
}
