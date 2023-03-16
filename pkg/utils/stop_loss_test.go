package utils

import (
	"testing"

	"github.com/c9s/bbgo/pkg/fixedpoint"
	"github.com/c9s/bbgo/pkg/types"
	"github.com/dop251/goja"
	"github.com/stretchr/testify/assert"
)

func TestParseStopLoss(t *testing.T) {
	vm := goja.New()
	val, err := ParseStopLoss(vm, types.SideTypeSell, fixedpoint.NewFromFloat(1.0), "dynamic")
	assert.NoError(t, err)
	assert.Nil(t, val)
}

func TestParseStopLossForSell(t *testing.T) {
	vm := goja.New()
	val, err := ParseStopLoss(vm, types.SideTypeSell, fixedpoint.NewFromFloat(1.0), "1%")
	assert.NoError(t, err)
	assert.Equal(t, 1.01, val.Float64())
}

func TestParseStopLossForBuy(t *testing.T) {
	vm := goja.New()
	val, err := ParseStopLoss(vm, types.SideTypeBuy, fixedpoint.NewFromFloat(1.0), "1%")
	assert.NoError(t, err)
	assert.Equal(t, 0.99, val.Float64())
}

func TestParseStopLossForFixed(t *testing.T) {
	vm := goja.New()
	val, err := ParseStopLoss(vm, types.SideTypeBuy, fixedpoint.NewFromFloat(1.0), "1.0*0.99")
	assert.NoError(t, err)
	assert.Equal(t, 0.99, val.Float64())
}
