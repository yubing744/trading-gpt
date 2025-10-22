package utils

import (
	"strings"

	"github.com/c9s/bbgo/pkg/fixedpoint"
	"github.com/c9s/bbgo/pkg/types"
	"github.com/dop251/goja"
)

// ParsePrice parses a price expression and returns a fixedpoint value
// Supports absolute values like "45000" or expressions like "last_close * 0.995"
func ParsePrice(vm *goja.Runtime, klines *types.KLineWindow, closePrice fixedpoint.Value, expr string) (*fixedpoint.Value, error) {
	if strings.TrimSpace(expr) == "" {
		return nil, nil
	}

	// Set up JavaScript context with available variables
	vm.Set("last_close", closePrice.Float64())
	vm.Set("close", closePrice.Float64())

	// Add recent kline data if available
	if klines != nil && klines.Len() > 0 {
		lastIdx := klines.Len() - 1
		lastKline := (*klines)[lastIdx]
		vm.Set("last_open", lastKline.Open.Float64())
		vm.Set("last_high", lastKline.High.Float64())
		vm.Set("last_low", lastKline.Low.Float64())
		vm.Set("last_volume", lastKline.Volume.Float64())

		// Add more historical data if available
		if klines.Len() > 1 {
			prevKline := (*klines)[lastIdx-1]
			vm.Set("prev_close", prevKline.Close.Float64())
		}
	}

	return ArgToFixedpoint(vm, expr)
}
