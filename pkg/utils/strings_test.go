package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJoinFloatSlice(t *testing.T) {
	assert.Equal(t, "0.0010 0.0300", JoinFloatSlice([]float64{0.001, 0.03}, " "))
}

func TestJoinFloatSlice2(t *testing.T) {
	assert.Equal(t, "0.0010 0.0300 0.0450", JoinFloatSlice([]float64{0.001, 0.03, 0.045}, " "))
}
