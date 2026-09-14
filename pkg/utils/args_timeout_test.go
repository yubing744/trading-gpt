package utils

import (
	"testing"
	"time"

	"github.com/dop251/goja"
	"github.com/stretchr/testify/assert"
)

// TestArgToFixedpoint_NormalExpression verifies that a normal expression still
// evaluates correctly with the default timeout in place (issue #92 must not
// change happy-path behaviour).
func TestArgToFixedpoint_NormalExpression(t *testing.T) {
	vm := goja.New()

	val, err := ArgToFixedpoint(vm, "1.22 * 4")
	assert.NoError(t, err)
	if assert.NotNil(t, val) {
		assert.InDelta(t, 4.88, val.Float64(), 1e-9)
	}
}

// TestArgToFixedpointWithTimeout_InterruptsRunawayScript verifies that an
// infinite-loop script is interrupted and returns an error rather than hanging
// the caller (issue #92).
func TestArgToFixedpointWithTimeout_InterruptsRunawayScript(t *testing.T) {
	vm := goja.New()

	start := time.Now()
	val, err := ArgToFixedpointWithTimeout(vm, "while (true) {}", 200*time.Millisecond)
	elapsed := time.Since(start)

	assert.Error(t, err, "runaway script must return an error")
	assert.Nil(t, val)
	// Must have returned promptly around the timeout, not hung.
	assert.Less(t, elapsed, 3*time.Second, "eval should be interrupted near the timeout, not hang")

	// The interrupt payload should surface in the error.
	assert.Contains(t, err.Error(), "execution timeout")
}

// TestArgToFixedpointWithTimeout_NormalUnaffected verifies that a fast, valid
// expression completes normally even with a short timeout configured.
func TestArgToFixedpointWithTimeout_NormalUnaffected(t *testing.T) {
	vm := goja.New()

	val, err := ArgToFixedpointWithTimeout(vm, "100 * 0.995", 5*time.Second)
	assert.NoError(t, err)
	if assert.NotNil(t, val) {
		assert.InDelta(t, 99.5, val.Float64(), 1e-9)
	}
}

// TestArgToFixedpointWithTimeout_RuntimeReusableAfterInterrupt verifies that the
// interrupt flag is cleared, so the same runtime can be reused after a timeout.
func TestArgToFixedpointWithTimeout_RuntimeReusableAfterInterrupt(t *testing.T) {
	vm := goja.New()

	_, err := ArgToFixedpointWithTimeout(vm, "while (true) {}", 100*time.Millisecond)
	assert.Error(t, err)

	// Reusing the same runtime must work (ClearInterrupt was called).
	val, err := ArgToFixedpointWithTimeout(vm, "1.5 + 5", 5*time.Second)
	assert.NoError(t, err)
	if assert.NotNil(t, val) {
		assert.InDelta(t, 6.5, val.Float64(), 1e-9)
	}
}
