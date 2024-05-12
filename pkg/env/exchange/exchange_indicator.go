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
	Name string
	Type config.IndicatorType
	Data interface{}
}

func NewExchangeIndicator(name string, cfg *config.IndicatorConfig, indicators *bbgo.StandardIndicatorSet) *ExchangeIndicator {
	indicator := &ExchangeIndicator{
		Name: name,
		Type: cfg.Type,
	}

	switch cfg.Type {
	case config.IndicatorTypeSMA:
		indicator.Data = indicators.SMA(types.IntervalWindow{
			Interval: cfg.GetInterval("interval", "5m"),
			Window:   cfg.GetInt("windowSize", 5),
		})
	case config.IndicatorTypeEWMA:
		indicator.Data = indicators.EWMA(types.IntervalWindow{
			Interval: cfg.GetInterval("interval", "5m"),
			Window:   cfg.GetInt("windowSize", 5),
		})
	case config.IndicatorTypeVWMA:
		indicator.Data = indicators.VWMA(types.IntervalWindow{
			Interval: cfg.GetInterval("interval", "5m"),
			Window:   cfg.GetInt("windowSize", 5),
		})
	case config.IndicatorTypeEMV:
		indicator.Data = indicators.EMV(types.IntervalWindow{
			Interval: cfg.GetInterval("interval", "5m"),
			Window:   cfg.GetInt("windowSize", 5),
		})
	case config.IndicatorTypeBOLL:
		indicator.Data = indicators.BOLL(types.IntervalWindow{
			Interval: cfg.GetInterval("interval", "5m"),
			Window:   cfg.GetInt("windowSize", 21),
		}, cfg.GetFloat("bandWidth", 2.0))
	case config.IndicatorTypeRSI:
		indicator.Data = indicators.RSI(types.IntervalWindow{
			Interval: cfg.GetInterval("interval", "5m"),
			Window:   cfg.GetInt("windowSize", 21),
		})
	default:
		log.Panic("not support type" + cfg.Type)
	}

	return indicator
}

func (ei *ExchangeIndicator) ToPrompts(maxWindowSize int) []string {
	switch ei.Type {
	case config.IndicatorTypeBOLL:
		return ei.BOLLToPrompts(ei.Data.(*indicator.BOLL), maxWindowSize)
	default:
		basicIndicator, ok := ei.Data.(IBasicIndicator)
		if ok {
			return ei.BasicToPrompts(ei.Name, ei.Type, basicIndicator, maxWindowSize)
		} else {
			log.Panic("not support type" + ei.Type)
		}
	}

	return []string{}
}

func (indicator *ExchangeIndicator) BOLLToPrompts(boll *indicator.BOLL, maxWindowSize int) []string {
	log.WithField("boll", boll).Info("handle BOLL values changed")

	upVals := boll.UpBand
	if len(upVals) > maxWindowSize {
		upVals = upVals[len(upVals)-maxWindowSize:]
	}

	midVals := boll.SMA.Values
	if len(midVals) > maxWindowSize {
		midVals = midVals[len(midVals)-maxWindowSize:]
	}

	downVals := boll.DownBand
	if len(downVals) > maxWindowSize {
		downVals = downVals[len(downVals)-maxWindowSize:]
	}

	sb := strings.Builder{}

	sb.WriteString("BOLL (Bollinger Bands) data changed:\n")
	sb.WriteString(fmt.Sprintf("# Data Recorded at %s Intervals\n", boll.Interval))
	sb.WriteString("# Column Meanings:\n")
	sb.WriteString("# Time:     Time Point Number, Starting from 0\n")
	sb.WriteString("# UpBand:   Upper Band Value\n")
	sb.WriteString("# SMA:      Simple Moving Average Value\n")
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

func (indicator *ExchangeIndicator) BasicToPrompts(name string, indicatorType config.IndicatorType, basicIndicator IBasicIndicator, maxWindowSize int) []string {
	log.
		WithField("name", name).
		WithField("indicatorType", indicatorType).
		WithField("maxWindowSize", maxWindowSize).
		Info("indicator values changed")

	vals := basicIndicatorToValues(basicIndicator)
	if len(vals) > maxWindowSize {
		vals = vals[:maxWindowSize]
	}

	msgs := make([]string, 0)

	if len(vals) > 0 {
		msgs = append(msgs, fmt.Sprintf("%s data changed: [%s], and the current %s at index 0 value is: %.3f",
			name,
			utils.JoinFloatSlice([]float64(vals), " "),
			name,
			basicIndicator.Last(0),
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
