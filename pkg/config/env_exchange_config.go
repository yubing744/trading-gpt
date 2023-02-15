package config

import (
	"github.com/c9s/bbgo/pkg/fixedpoint"
	"github.com/c9s/bbgo/pkg/types"
)

type EnvExchangeConfig struct {
	Symbol   string         `json:"symbol"`
	Interval types.Interval `json:"interval"`
	// Leverage uses the account net value to calculate the order qty
	Leverage fixedpoint.Value `json:"leverage"`
	// Quantity sets the fixed order qty, takes precedence over Leverage
	Quantity fixedpoint.Value `json:"quantity"`
	// The tread threshold
	TrendThreshold fixedpoint.Value `json:"trend_threshold"`
}
