package exchange

import (
	"fmt"

	"github.com/c9s/bbgo/pkg/bbgo"
	"github.com/c9s/bbgo/pkg/indicator"
	"github.com/c9s/bbgo/pkg/types"

	"github.com/yubing744/trading-gpt/pkg/config"
	"github.com/yubing744/trading-gpt/pkg/utils"
)

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
	case config.IndicatorTypeRSI:
		return ei.RSIToPrompts(ei.Data.(*indicator.RSI), maxWindowSize)
	default:
		log.Panic("not support type" + ei.Type)
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

	tip1 := fmt.Sprintf("BOLL data changed: UpBand:[%s], SMA:[%s], DownBand:[%s]",
		utils.JoinFloatSlice([]float64(upVals), " "),
		utils.JoinFloatSlice([]float64(midVals), " "),
		utils.JoinFloatSlice([]float64(downVals), " "),
	)

	tip2 := fmt.Sprintf("The current UpBand is %.3f, and the current SMA is %.3f, and the current DownBand is %.3f",
		boll.UpBand.Last(0),
		boll.SMA.Last(0),
		boll.DownBand.Last(0),
	)

	return []string{tip1, tip2}
}

func (indicator *ExchangeIndicator) RSIToPrompts(rsi *indicator.RSI, maxWindowSize int) []string {
	log.WithField("rsi", rsi).Info("handle RSI values changed")

	vals := rsi.Values
	if len(vals) > maxWindowSize {
		vals = vals[len(vals)-maxWindowSize:]
	}

	msg := fmt.Sprintf("RSI data changed: [%s], and the current RSI value is: %.3f",
		utils.JoinFloatSlice([]float64(vals), " "),
		rsi.Last(0),
	)

	return []string{msg}
}
