package config

import (
	"strconv"
	"time"

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
}
