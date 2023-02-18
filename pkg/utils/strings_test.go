package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJoinFloatSlice(t *testing.T) {
	assert.Equal(t, "0.001 0.030", JoinFloatSlice([]float64{0.001, 0.03}, " "))
}

func TestJoinFloatSlice2(t *testing.T) {
	assert.Equal(t, "0.001 0.030 0.045", JoinFloatSlice([]float64{0.001, 0.03, 0.045}, " "))
}
