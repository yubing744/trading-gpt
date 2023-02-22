package utils

import (
	"fmt"
	"strings"

	"github.com/c9s/bbgo/pkg/types"
)

func JoinFloatSlice[T float64 | float32](data []T, sep string) string {
	var builder strings.Builder

	total := len(data)
	for i, item := range data {
		builder.WriteString(fmt.Sprintf("%.3f", item))

		if i < total-1 {
			builder.WriteString(sep)
		}
	}

	return builder.String()
}

func JoinFloatSeries(data types.Series, sep string) string {
	var builder strings.Builder

	total := data.Length()
	for i := total - 1; i >= 0; i-- {
		builder.WriteString(fmt.Sprintf("%.3f", data.Index(i)))

		if i > 0 {
			builder.WriteString(sep)
		}
	}

	return builder.String()
}

func Contains(arr []string, item string) bool {
	for _, ele := range arr {
		if ele == item {
			return true
		}
	}

	return false
}
