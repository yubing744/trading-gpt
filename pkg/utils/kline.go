package utils

import (
	"fmt"
	"strings"

	"github.com/c9s/bbgo/pkg/types"
)

func FormatKLineWindow(window types.KLineWindow) string {
	var sb strings.Builder

	// Add descriptive information
	sb.WriteString(fmt.Sprintf("# Data Recorded at %s Candlestick Interval\n", window.GetInterval()))
	sb.WriteString("# Column Meanings:\n")
	sb.WriteString("# Time:  Candlestick Period Number, Starting from 0\n")
	sb.WriteString("\n")

	// Add header
	sb.WriteString("Time   Open   Close   High    Low     Volume\n")

	// Iterate through candlestick data and format as rows
	for i, kline := range window {
		sb.WriteString(fmt.Sprintf("%d      %.3f  %.3f    %.3f   %.3f   %.3f\n", i, kline.Open.Float64(), kline.Close.Float64(), kline.High.Float64(), kline.Low.Float64(), kline.Volume.Float64()))
	}

	// Add latest closing price
	sb.WriteString(fmt.Sprintf("\nCurrent close price: %.3f", window.Close().Last(0)))

	return sb.String()
}
