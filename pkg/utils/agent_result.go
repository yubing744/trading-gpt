package utils

import (
	"encoding/json"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"github.com/yubing744/trading-gpt/pkg/types"
)

func extractThinking(text string) string {
	return ""
}

func ParseResult(text string) (*types.Result, error) {
	text = trimJSON(trimMarkdownJSON(strings.ReplaceAll(text, "\\", "")))
	jsonBytes := removeJSONComments([]byte(text))

	var result types.Result
	err := json.Unmarshal(jsonBytes, &result)
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
			return text[jsonStart : jsonEnd+1]
		}
	}

	return text
}

func removeJSONComments(jsonData []byte) []byte {
	singleLineCommentPattern := regexp.MustCompile(`//.*$`)
	jsonData = singleLineCommentPattern.ReplaceAll(jsonData, []byte{})

	multiLineCommentPattern := regexp.MustCompile(`/\*.*?\*/`)
	jsonData = multiLineCommentPattern.ReplaceAll(jsonData, []byte{})
	return jsonData
}
