package utils

import (
	"fmt"
	"strings"

	"github.com/c9s/bbgo/pkg/fixedpoint"
	"github.com/c9s/bbgo/pkg/types"
)

func FormatKLineWindow(window types.KLineWindow, maxNum int) string {
	var sb strings.Builder

	// Add descriptive information
	sb.WriteString(fmt.Sprintf("# Data Recorded at %s Candlestick Interval\n", window.GetInterval()))
	sb.WriteString("# Column Meanings:\n")
	sb.WriteString("# Time:  Candlestick Period Number, Starting from 0\n")
	sb.WriteString("\n")

	// Add header
	sb.WriteString("Time   Open   Close   High    Low     Volume    Change    Change%    Amplitude\n")

	if len(window) == 0 {
		return sb.String()
	}

	// Iterate through candlestick data and format as rows
	for i, kline := range window.Tail(maxNum) {
		change := kline.GetChange()
		changePercent := change.Div(kline.GetOpen()).Mul(fixedpoint.NewFromFloat(100))
		amplitude := kline.GetMaxChange().Div(kline.GetLow()).Mul(fixedpoint.NewFromFloat(100))
		sb.WriteString(fmt.Sprintf("%d      %.3f  %.3f    %.3f   %.3f   %.3f    %.3f    %.3f%%    %.3f%%\n", i, kline.Open.Float64(), kline.Close.Float64(), kline.High.Float64(), kline.Low.Float64(), kline.Volume.Float64(), change.Float64(), changePercent.Float64(), amplitude.Float64()))
	}

	// Add latest closing price
	sb.WriteString(fmt.Sprintf("\nCurrent close price: %.3f", window.Close().Last(0)))

	return sb.String()
}
