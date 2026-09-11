package config

import (
	"strconv"
	"time"

	"github.com/c9s/bbgo/pkg/fixedpoint"
	"github.com/c9s/bbgo/pkg/types"
)

type IndicatorType string

const (
	IndicatorTypeSMA          IndicatorType = "sma"
	IndicatorTypeEWMA         IndicatorType = "ewma"
	IndicatorTypeVWMA         IndicatorType = "vwma"
	IndicatorTypePivotHigh    IndicatorType = "pivothigh"
	IndicatorTypePivotLow     IndicatorType = "pivotlow"
	IndicatorTypeATR          IndicatorType = "atr"
	IndicatorTypeATRP         IndicatorType = "atrp"
	IndicatorTypeEMV          IndicatorType = "emv"
	IndicatorTypeCCI          IndicatorType = "cci"
	IndicatorTypeHULL         IndicatorType = "hull"
	IndicatorTypeSTOCH        IndicatorType = "stoch"
	IndicatorTypeBOLL         IndicatorType = "boll"
	IndicatorTypeMACDLegacy   IndicatorType = "macdlegacy"
	IndicatorTypeRSI          IndicatorType = "rsi"
	IndicatorTypeGHFilter     IndicatorType = "ghfilter"
	IndicatorTypeKalmanFilter IndicatorType = "kalmanfilter"
	IndicatorTypeVR           IndicatorType = "vr"
)

type IndicatorConfig struct {
	Type   IndicatorType     `json:"type"`
	MaxNum *int              `json:"max_num"`
	Params map[string]string `json:"params"`
}

func (cfg IndicatorConfig) GetString(key string, def string) string {
	val, ok := cfg.Params[key]
	if !ok {
		return def
	}

	return val
}

func (cfg IndicatorConfig) GetInt(key string, def int) int {
	val, ok := cfg.Params[key]
	if !ok {
		return def
	}

	intVal, err := strconv.Atoi(val)
	if err != nil {
		return def
	}

	return intVal
}

func (cfg IndicatorConfig) GetFloat(key string, def float64) float64 {
	val, ok := cfg.Params[key]
	if !ok {
		return def
	}

	floatValue, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return def
	}

	return floatValue
}

func (cfg IndicatorConfig) GetDuration(key string, def types.Duration) types.Duration {
	val, ok := cfg.Params[key]
	if !ok {
		return def
	}

	duration, err := time.ParseDuration(val)
	if err != nil {
		return def
	}

	return types.Duration(duration)
}

func (cfg IndicatorConfig) GetInterval(key string, def types.Interval) types.Interval {
	val, ok := cfg.Params[key]
	if !ok {
		return def
	}

	return types.Interval(val)
}

type EnvExchangeConfig struct {
	KlineNum            int                         `json:"kline_num"`
	Indicators          map[string]*IndicatorConfig `json:"indicators"`
	HandlePositionClose bool                        `json:"handle_position_close"`
	CleanPosition       CleanPositionConfig         `json:"clean_position"`
	RiskControl         RiskControlConfig           `json:"risk_control"`
}

type CleanPositionConfig struct {
	Enabled  bool           `json:"enabled"`
	Interval types.Interval `json:"interval"`
}

// RiskControlConfig defines code-level hard risk-control gates that operate
// independently of the LLM decision. Any configured limit that is breached
// causes the corresponding order submission to be rejected (defensive deny),
// never a new/active order. See issue #86 (KR3).
//
// Threshold semantics: a zero (unset) limit disables that individual check,
// so an unconfigured deployment keeps its exact prior behaviour. Operators
// tighten the gate by setting positive thresholds (recommended, sized to the
// live account). The master `enabled` switch defaults to true.
type RiskControlConfig struct {
	// Enabled is the master switch for all hard risk gates. Pointer so that an
	// absent config field defaults to enabled (true); set `enabled: false` to
	// explicitly disable. Individual checks are still no-ops until their
	// thresholds are set to positive values.
	Enabled *bool `json:"enabled"`

	// MaxLeverage caps the effective leverage used to open/add a position.
	// Orders whose leverage exceeds this are rejected. Zero disables the check.
	MaxLeverage fixedpoint.Value `json:"max_leverage"`

	// MaxOrderQuote caps the notional (in quote currency) of a single open/add
	// order. Zero disables the check.
	MaxOrderQuote fixedpoint.Value `json:"max_order_quote"`

	// MaxPositionQuote caps the total notional (in quote currency) of the
	// resulting position (existing exposure + this order). Zero disables it.
	MaxPositionQuote fixedpoint.Value `json:"max_position_quote"`

	// MaxDailyLoss is the absolute realized loss (quote currency, positive
	// number) that trips the daily kill-switch. Once tripped, all new
	// open/add orders are rejected for the remainder of the trading day. The
	// state resets automatically at the next trading day. Zero disables it.
	MaxDailyLoss fixedpoint.Value `json:"max_daily_loss"`

	// AutoCloseOnKill, when true, would close open positions on kill-switch
	// trip. This is an ACTIVE fund action and is intentionally NOT performed
	// autonomously by this code (requires owner approval); the flag is kept
	// for forward compatibility and only emits an alert. Default false.
	AutoCloseOnKill bool `json:"auto_close_on_kill"`
}

// IsEnabled reports whether the hard risk gate is active. Defaults to true
// when the config omits the `enabled` field.
func (c RiskControlConfig) IsEnabled() bool {
	if c.Enabled == nil {
		return true
	}
	return *c.Enabled
}
