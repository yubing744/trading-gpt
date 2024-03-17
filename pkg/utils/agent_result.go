package utils

import (
	"encoding/json"
	"strings"

	"github.com/yubing744/trading-gpt/pkg/types"
)

func ParseResult(text string) (*types.Result, error) {
	jsonStart := strings.Index(text, "```json")

	if jsonStart >= 0 {
		jsonEnd := strings.LastIndex(text, "```")

		if jsonEnd >= 0 {
			text = text[jsonStart+7 : jsonEnd]
		}
	}

	var result types.Result
	err := json.Unmarshal([]byte(text), &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
