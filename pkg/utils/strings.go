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
		builder.WriteString(fmt.Sprintf("%.4f", item))

		if i < total-1 {
			builder.WriteString(sep)
		}
	}

	return builder.String()
}

func JoinFloatSeries(data types.Series, sep string) string {
	var builder strings.Builder

	total := data.Length()
	for i := 0; i < total; i++ {
		builder.WriteString(fmt.Sprintf("%.4f", data.Index(i)))

		if i < total-1 {
			builder.WriteString(sep)
		}
	}

	return builder.String()
}
