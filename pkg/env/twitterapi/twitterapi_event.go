package twitterapi

import (
	"fmt"
	"strings"

	"github.com/yubing744/trading-gpt/pkg/types"
)

// TwitterAPIEvent represents an event that is specific to interactions with the Twitter API platform.
type TwitterAPIEvent struct {
	types.Event // Embed the base Event struct to reuse its implementation.
	title       string
	Content     string
}

// NewTwitterAPIEvent creates a new instance of TwitterAPIEvent with the given type and data.
func NewTwitterAPIEvent(name string, title string, content string) *TwitterAPIEvent {
	return &TwitterAPIEvent{
		Event:   *types.NewEvent(name, content),
		title:   title,
		Content: content,
	}
}

// ToPrompts is overridden to include the message content in the prompts for TwitterAPIEvent.
func (e *TwitterAPIEvent) ToPrompts() []string {
	sb := strings.Builder{}

	sb.WriteString(fmt.Sprintf("%s\n", e.title))
	sb.WriteString(e.Content)

	return []string{sb.String()}
}
