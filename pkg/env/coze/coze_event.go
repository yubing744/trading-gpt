package coze

import (
	"fmt"
	"strings"

	"github.com/yubing744/trading-gpt/pkg/types"
)

// CozeEvent represents an event that is specific to interactions with the Coze platform.
type CozeEvent struct {
	types.Event // Embed the base Event struct to reuse its implementation.
	title       string
	Content     string
}

// NewCozeEvent creates a new instance of CozeEvent with the given type and data.
func NewCozeEvent(name string, title string, content string) *CozeEvent {
	return &CozeEvent{
		Event:   *types.NewEvent(name, content),
		title:   title,
		Content: content,
	}
}

// ToPrompts is overridden to include the message content in the prompts for CozeEvent.
func (e *CozeEvent) ToPrompts() []string {
	sb := strings.Builder{}

	sb.WriteString(fmt.Sprintf("%s\n", e.title))
	sb.WriteString(e.Content)

	return []string{sb.String()}
}
