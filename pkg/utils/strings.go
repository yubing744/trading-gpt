package utils

import (
	"fmt"
	"strings"
)

func JoinFloatSlice[T float64 | float32](data []T, sep string) string {
	var builder strings.Builder

	total := len(data)
	for i, item := range data {
		builder.WriteString(fmt.Sprintf("%.2f", item))

		if i < total-1 {
			builder.WriteString(sep)
		}
	}

	return builder.String()
}
