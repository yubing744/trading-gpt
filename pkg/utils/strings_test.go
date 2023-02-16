package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJoinFloatSlice(t *testing.T) {
	assert.Equal(t, "0.00 0.03", JoinFloatSlice([]float64{0.001, 0.03}, " "))
}

func TestJoinFloatSlice2(t *testing.T) {
	assert.Equal(t, "0.00 0.03 0.04", JoinFloatSlice([]float64{0.001, 0.03, 0.045}, " "))
}
