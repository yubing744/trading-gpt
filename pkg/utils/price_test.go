package utils

import (
	"testing"

	"github.com/c9s/bbgo/pkg/fixedpoint"
	"github.com/c9s/bbgo/pkg/types"
	"github.com/dop251/goja"
	"github.com/stretchr/testify/assert"
)

func TestParsePrice_AbsoluteValue(t *testing.T) {
	vm := goja.New()
	closePrice := fixedpoint.NewFromFloat(100.0)

	price, err := ParsePrice(vm, nil, closePrice, "95.5")
	assert.NoError(t, err)
	assert.NotNil(t, price)
	assert.Equal(t, 95.5, price.Float64())
}

func TestParsePrice_Expression(t *testing.T) {
	vm := goja.New()
	closePrice := fixedpoint.NewFromFloat(100.0)

	price, err := ParsePrice(vm, nil, closePrice, "last_close * 0.995")
	assert.NoError(t, err)
	assert.NotNil(t, price)
	assert.InDelta(t, 99.5, price.Float64(), 0.01)
}

func TestParsePrice_CloseAlias(t *testing.T) {
	vm := goja.New()
	closePrice := fixedpoint.NewFromFloat(100.0)

	price, err := ParsePrice(vm, nil, closePrice, "close * 1.005")
	assert.NoError(t, err)
	assert.NotNil(t, price)
	assert.InDelta(t, 100.5, price.Float64(), 0.01)
}

func TestParsePrice_WithKlines(t *testing.T) {
	vm := goja.New()
	closePrice := fixedpoint.NewFromFloat(100.0)

	klines := &types.KLineWindow{
		{
			Open:   fixedpoint.NewFromFloat(98.0),
			High:   fixedpoint.NewFromFloat(105.0),
			Low:    fixedpoint.NewFromFloat(95.0),
			Close:  fixedpoint.NewFromFloat(100.0),
			Volume: fixedpoint.NewFromFloat(1000.0),
		},
	}

	price, err := ParsePrice(vm, klines, closePrice, "last_low * 1.01")
	assert.NoError(t, err)
	assert.NotNil(t, price)
	assert.InDelta(t, 95.95, price.Float64(), 0.01)
}

func TestParsePrice_WithHistoricalKlines(t *testing.T) {
	vm := goja.New()
	closePrice := fixedpoint.NewFromFloat(100.0)

	klines := &types.KLineWindow{
		{
			Close: fixedpoint.NewFromFloat(98.0),
		},
		{
			Close: fixedpoint.NewFromFloat(100.0),
		},
	}

	price, err := ParsePrice(vm, klines, closePrice, "prev_close * 0.99")
	assert.NoError(t, err)
	assert.NotNil(t, price)
	assert.InDelta(t, 97.02, price.Float64(), 0.01)
}

func TestParsePrice_UsingHighAndLow(t *testing.T) {
	vm := goja.New()
	closePrice := fixedpoint.NewFromFloat(100.0)

	klines := &types.KLineWindow{
		{
			High: fixedpoint.NewFromFloat(105.0),
			Low:  fixedpoint.NewFromFloat(95.0),
		},
	}

	// Test using last_high
	price, err := ParsePrice(vm, klines, closePrice, "last_high * 0.95")
	assert.NoError(t, err)
	if assert.NotNil(t, price) {
		assert.InDelta(t, 99.75, price.Float64(), 0.01)
	}

	// Test using last_low
	vm2 := goja.New()
	price2, err2 := ParsePrice(vm2, klines, closePrice, "last_low * 1.05")
	assert.NoError(t, err2)
	if assert.NotNil(t, price2) {
		assert.InDelta(t, 99.75, price2.Float64(), 0.01)
	}
}

func TestParsePrice_EmptyString(t *testing.T) {
	vm := goja.New()
	closePrice := fixedpoint.NewFromFloat(100.0)

	price, err := ParsePrice(vm, nil, closePrice, "")
	assert.NoError(t, err)
	assert.Nil(t, price)
}

func TestParsePrice_WhitespaceString(t *testing.T) {
	vm := goja.New()
	closePrice := fixedpoint.NewFromFloat(100.0)

	price, err := ParsePrice(vm, nil, closePrice, "   ")
	assert.NoError(t, err)
	assert.Nil(t, price)
}

func TestParsePrice_InvalidExpression(t *testing.T) {
	vm := goja.New()
	closePrice := fixedpoint.NewFromFloat(100.0)

	price, err := ParsePrice(vm, nil, closePrice, "invalid syntax +++")
	assert.Error(t, err)
	assert.Nil(t, price)
}
