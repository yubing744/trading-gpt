package env

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewEnvironment(t *testing.T) {
	env := NewEnvironment()
	assert.NotNil(t, env)
}
