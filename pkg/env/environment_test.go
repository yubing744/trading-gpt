package env

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yubing744/trading-bot/pkg/config"
)

func TestNewEnvironment(t *testing.T) {
	env := NewEnvironment(&config.EnvConfig{})
	assert.NotNil(t, env)
}
