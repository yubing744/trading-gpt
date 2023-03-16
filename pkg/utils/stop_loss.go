package utils

import (
	"strings"

	"github.com/c9s/bbgo/pkg/fixedpoint"
	"github.com/c9s/bbgo/pkg/types"
	"github.com/dop251/goja"
)

func ParseStopLoss(vm *goja.Runtime, side types.SideType, closePrice fixedpoint.Value, text string) (*fixedpoint.Value, error) {
	if strings.Contains(text, "dynamic") {
		return nil, nil
	}

	if strings.Contains(text, "%") {
		val, err := fixedpoint.NewFromString(text)
		if err != nil {
			return nil, err
		}

		switch side {
		case types.SideTypeSell:
			val = closePrice.Mul(fixedpoint.One.Add(val))
			return &val, nil
		case types.SideTypeBuy:
			val = closePrice.Mul(fixedpoint.One.Sub(val))
			return &val, nil
		default:
			return nil, nil
		}
	}

	return ArgToFixedpoint(vm, text)
}
