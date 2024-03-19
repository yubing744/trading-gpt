package utils

import (
	"encoding/json"
	"strings"

	"github.com/pkg/errors"
	"github.com/yubing744/trading-gpt/pkg/types"
)

func ParseResult(text string) (*types.Result, error) {
	text = trimJSON(trimMarkdownJSON(strings.ReplaceAll(text, "\\", "")))

	var result types.Result
	err := json.Unmarshal([]byte(text), &result)
	if err != nil {
		return nil, errors.Wrap(err, "json.Unmarshal_error")
	}

	return &result, nil
}

func trimMarkdownJSON(text string) string {
	jsonStart := strings.Index(text, "```json")

	if jsonStart >= 0 {
		jsonEnd := strings.LastIndex(text, "```")

		if jsonEnd >= 0 {
			text = text[jsonStart+7 : jsonEnd]
		}
	}

	return text
}

func trimJSON(text string) string {
	jsonStart := strings.Index(text, "{")

	if jsonStart >= 0 {
		jsonEnd := strings.LastIndex(text, "}")

		if jsonEnd >= 0 {
			text = text[jsonStart : jsonEnd+1]
		}
	}

	return text
}
