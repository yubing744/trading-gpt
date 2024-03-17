package utils

import (
	"encoding/json"

	"github.com/yubing744/trading-gpt/pkg/types"
)

func ParseResult(text string) (*types.Result, error) {
	var result types.Result
	err := json.Unmarshal([]byte(text), &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
