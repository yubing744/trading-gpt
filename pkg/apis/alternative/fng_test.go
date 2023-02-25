package alternative

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetFearAndGreedIndex(t *testing.T) {
	an := NewAlternativeClient()
	assert.NotNil(t, an)

	index, err := an.GetFearAndGreedIndex(1)
	assert.NoError(t, err)
	assert.NotNil(t, index)

	assert.Equal(t, "Fear and Greed Index", index.Name)
	assert.Len(t, index.Data, 1)
}
