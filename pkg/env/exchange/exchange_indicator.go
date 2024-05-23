package exchange

import (
	"fmt"
	"strings"

	"github.com/c9s/bbgo/pkg/bbgo"
	"github.com/c9s/bbgo/pkg/indicator"
	"github.com/c9s/bbgo/pkg/types"

	"github.com/yubing744/trading-gpt/pkg/config"
	"github.com/yubing744/trading-gpt/pkg/utils"
)

type IBasicIndicator interface {
	Length() int
	Index(i int) float64
	Last(i int) float64
}

type ExchangeIndicator struct {
	Name   string
	Type   config.IndicatorType
	Config *config.IndicatorConfig
	Data   interface{}
}

func NewExchangeIndicator(name string, cfg *config.IndicatorConfig, indicators *bbgo.StandardIndicatorSet) *ExchangeIndicator {
	indicator := &ExchangeIndicator{
		Name:   name,
		Type:   cfg.Type,
		Config: cfg,
	}

	switch cfg.Type {
	case config.IndicatorTypeSMA:
		indicator.Data = indicators.SMA(types.IntervalWindow{
			Interval: cfg.GetInterval("interval", "5m"),
			Window:   cfg.GetInt("window_size", 5),
		})
	case config.IndicatorTypeVR:
		indicator.Data = indicators.VR(types.IntervalWindow{
			Interval: cfg.GetInterval("interval", "5m"),
			Window:   cfg.GetInt("window_size", 5),
		})
	case config.IndicatorTypeEWMA:
		indicator.Data = indicators.EWMA(types.IntervalWindow{
			Interval: cfg.GetInterval("interval", "5m"),
			Window:   cfg.GetInt("window_size", 5),
		})
	case config.IndicatorTypeVWMA:
		indicator.Data = indicators.VWMA(types.IntervalWindow{
			Interval: cfg.GetInterval("interval", "5m"),
			Window:   cfg.GetInt("window_size", 5),
		})
	case config.IndicatorTypeEMV:
		indicator.Data = indicators.EMV(types.IntervalWindow{
			Interval: cfg.GetInterval("interval", "5m"),
			Window:   cfg.GetInt("window_size", 5),
		})
	case config.IndicatorTypeBOLL:
		indicator.Data = indicators.BOLL(types.IntervalWindow{
			Interval: cfg.GetInterval("interval", "5m"),
			Window:   cfg.GetInt("window_size", 20),
		}, cfg.GetFloat("band_width", 2.0))
	case config.IndicatorTypeRSI:
		indicator.Data = indicators.RSI(types.IntervalWindow{
			Interval: cfg.GetInterval("interval", "5m"),
			Window:   cfg.GetInt("window_size", 20),
		})
	case config.IndicatorTypeATR:
		indicator.Data = indicators.ATR(types.IntervalWindow{
			Interval: cfg.GetInterval("interval", "5m"),
			Window:   cfg.GetInt("window_size", 20),
		})
	case config.IndicatorTypeATRP:
		indicator.Data = indicators.ATRP(types.IntervalWindow{
			Interval: cfg.GetInterval("interval", "5m"),
			Window:   cfg.GetInt("window_size", 20),
		})
	default:
		log.Panic("not support type" + cfg.Type)
	}

	return indicator
}

func (ei *ExchangeIndicator) ToPrompts(maxNum int) []string {
	if ei.Config.MaxNum != nil {
		maxNum = *ei.Config.MaxNum
	}

	switch ei.Type {
	case config.IndicatorTypeBOLL:
		return ei.BOLLToPrompts(ei.Name, ei.Type, ei.Data.(*indicator.BOLL), maxNum)
	default:
		basicIndicator, ok := ei.Data.(IBasicIndicator)
		if ok {
			return ei.BasicToPrompts(ei.Name, ei.Type, basicIndicator, maxNum)
		} else {
			log.Panic("not support type" + ei.Type)
		}
	}

	return []string{}
}

func (indicator *ExchangeIndicator) BOLLToPrompts(name string, indicatorType config.IndicatorType, boll *indicator.BOLL, maxNum int) []string {
	log.
		WithField("name", name).
		WithField("indicatorType", indicatorType).
		WithField("maxWindowSize", maxNum).
		Info("handle BOLL values changed")

	upVals := boll.UpBand
	if len(upVals) > maxNum {
		upVals = upVals[len(upVals)-maxNum:]
	}

	midVals := boll.SMA.Values
	if len(midVals) > maxNum {
		midVals = midVals[len(midVals)-maxNum:]
	}

	downVals := boll.DownBand
	if len(downVals) > maxNum {
		downVals = downVals[len(downVals)-maxNum:]
	}

	sb := strings.Builder{}

	sb.WriteString(fmt.Sprintf("%s (Bollinger Bands) data changed:\n", name))
	sb.WriteString("# Column Meanings:\n")
	sb.WriteString("# Time:     Time Point Number, Starting from 0\n")
	sb.WriteString("# UpBand:   Upper Band Value\n")
	sb.WriteString(fmt.Sprintf("# SMA:      Simple Moving Average Value by %d windowSize\n", boll.SMA.Window))
	sb.WriteString("# DownBand: Lower Band Value\n")
	sb.WriteString("\n")

	sb.WriteString("Time   UpBand   SMA   DownBand\n")
	for i := 0; i < len(upVals); i++ {
		sb.WriteString(fmt.Sprintf("%d      %.3f  %.3f    %.3f\n", i, upVals[i], midVals[i], downVals[i]))
	}

	sb.WriteString("\n")

	sb.WriteString(fmt.Sprintf("The current UpBand is %.3f, and the current SMA is %.3f, and the current DownBand is %.3f",
		boll.UpBand.Last(0),
		boll.SMA.Last(0),
		boll.DownBand.Last(0),
	))

	return []string{sb.String()}
}

func (indicator *ExchangeIndicator) BasicToPrompts(name string, indicatorType config.IndicatorType, basicIndicator IBasicIndicator, maxNum int) []string {
	log.
		WithField("name", name).
		WithField("indicatorType", indicatorType).
		WithField("maxWindowSize", maxNum).
		Info("indicator values changed")

	vals := basicIndicatorToValues(basicIndicator)
	if len(vals) > maxNum {
		vals = vals[:maxNum]
	}

	// Reverse the vals slice
	for i, j := 0, len(vals)-1; i < j; i, j = i+1, j-1 {
		vals[i], vals[j] = vals[j], vals[i]
	}

	msgs := make([]string, 0)

	if len(vals) > 0 {
		msgs = append(msgs, fmt.Sprintf("%s data changed: [%s], and the most recent %s value is: %.3f at index %d",
			name,
			utils.JoinFloatSlice(vals, " "),
			name,
			basicIndicator.Last(0),
			len(vals)-1,
		))
	}

	return msgs
}

func basicIndicatorToValues(basicIndicator IBasicIndicator) []float64 {
	vals := make([]float64, 0)

	for i := 0; i < basicIndicator.Length(); i++ {
		vals = append(vals, basicIndicator.Index(i))
	}

	return vals
}
