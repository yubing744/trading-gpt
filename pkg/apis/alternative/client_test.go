package alternative

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewAlternativeClient(t *testing.T) {
	an := NewAlternativeClient()
	assert.NotNil(t, an)
}
