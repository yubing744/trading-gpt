package openai

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewOpenAIAgent(t *testing.T) {
	agent := NewOpenAIAgent()
	assert.NotNil(t, agent)
}
