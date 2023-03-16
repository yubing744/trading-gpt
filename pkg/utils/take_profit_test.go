package utils

import (
	"testing"

	"github.com/c9s/bbgo/pkg/fixedpoint"
	"github.com/c9s/bbgo/pkg/types"
	"github.com/dop251/goja"
	"github.com/stretchr/testify/assert"
)

func TestTakeProfitLoss(t *testing.T) {
	vm := goja.New()
	val, err := ParseTakeProfit(vm, types.SideTypeSell, fixedpoint.NewFromFloat(1.0), "dynamic")
	assert.NoError(t, err)
	assert.Nil(t, val)
}

func TestParseTakeProfitForSell(t *testing.T) {
	vm := goja.New()
	val, err := ParseTakeProfit(vm, types.SideTypeSell, fixedpoint.NewFromFloat(1.0), "10%")
	assert.NoError(t, err)
	assert.Equal(t, 0.9, val.Float64())
}

func TestParseTakeProfitForBuy(t *testing.T) {
	vm := goja.New()
	val, err := ParseTakeProfit(vm, types.SideTypeBuy, fixedpoint.NewFromFloat(1.0), "10%")
	assert.NoError(t, err)
	assert.Equal(t, 1.1, val.Float64())
}

func TestParseTakeProfitForFixed(t *testing.T) {
	vm := goja.New()
	val, err := ParseTakeProfit(vm, types.SideTypeBuy, fixedpoint.NewFromFloat(1.0), "1.0*0.99")
	assert.NoError(t, err)
	assert.Equal(t, 0.99, val.Float64())
}
