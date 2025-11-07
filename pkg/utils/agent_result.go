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
	// Try parsing with increasing levels of repair
	repairs := []func(string) string{
		// Level 1: Basic cleanup
		func(s string) string {
			s = trimMarkdownJSON(s)
			s = trimJSON(s)
			return s
		},
		// Level 2: Fix escaped underscores and backslashes
		func(s string) string {
			s = trimMarkdownJSON(s)
			s = fixEscapedUnderscores(s)
			s = trimJSON(s)
			return s
		},
		// Level 3: More aggressive cleanup
		func(s string) string {
			s = trimMarkdownJSON(s)
			s = fixEscapedUnderscores(s)
			s = fixCommonJSONIssues(s)
			s = trimJSON(s)
			return s
		},
	}

	var lastErr error
	for i, repair := range repairs {
		cleanedText := repair(text)
		jsonBytes := removeJSONComments([]byte(cleanedText))

		var result types.Result
		err := json.Unmarshal(jsonBytes, &result)
		if err == nil {
			return &result, nil
		}

		lastErr = err
		// Log attempt if not the last one
		if i < len(repairs)-1 {
			continue
		}
	}

	return nil, errors.Wrapf(lastErr, "json.Unmarshal_error after all repair attempts")
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

// fixEscapedUnderscores fixes improperly escaped underscores in JSON strings
// Common issue: LLMs sometimes escape underscores like "open\_long\_position"
func fixEscapedUnderscores(text string) string {
	// Fix escaped underscores in JSON keys and values
	text = strings.ReplaceAll(text, "\\_", "_")
	return text
}

// fixCommonJSONIssues applies various fixes for common JSON formatting issues
func fixCommonJSONIssues(text string) string {
	// Remove extra whitespace and newlines
	text = strings.TrimSpace(text)

	// Fix trailing commas before closing braces/brackets
	trailingCommaPattern := regexp.MustCompile(`,\s*([}\]])`)
	text = trailingCommaPattern.ReplaceAllString(text, "$1")

	// Fix missing commas between object/array elements
	// Pattern 1: } followed by whitespace and { (between objects in array)
	text = regexp.MustCompile(`}\s*\n\s*{`).ReplaceAllString(text, "},\n{")

	// Pattern 2: } followed by whitespace and " (between object and next key)
	text = regexp.MustCompile(`}\s*\n\s*"`).ReplaceAllString(text, "},\n\"")

	// Pattern 3: "value" followed by newline and "key": (missing comma between properties)
	text = regexp.MustCompile(`"([^"]*?)"\s*\n\s*"([^"]*?)"\s*:`).ReplaceAllString(text, `"$1",\n"$2":`)

	// Fix single quotes to double quotes (JSON only allows double quotes)
	// Be careful not to replace apostrophes inside strings
	text = fixSingleQuotes(text)

	// Fix unescaped newlines in string values
	text = fixUnescapedNewlines(text)

	// Remove any non-JSON prefix/suffix text more aggressively
	text = extractJSONObject(text)

	return text
}

// fixSingleQuotes converts single quotes to double quotes for JSON keys
func fixSingleQuotes(text string) string {
	// Pattern: 'key': value -> "key": value
	singleQuoteKeyPattern := regexp.MustCompile(`'([^']+)'\s*:`)
	text = singleQuoteKeyPattern.ReplaceAllString(text, `"$1":`)
	return text
}

// fixUnescapedNewlines fixes unescaped newlines within JSON strings
func fixUnescapedNewlines(text string) string {
	var (
		builder  strings.Builder
		inString bool
		escape   bool
	)

	builder.Grow(len(text))

	for i := 0; i < len(text); i++ {
		ch := text[i]

		if escape {
			builder.WriteByte(ch)
			escape = false
			continue
		}

		switch ch {
		case '\\':
			escape = true
			builder.WriteByte(ch)
		case '"':
			inString = !inString
			builder.WriteByte(ch)
		case '\r':
			if inString {
				// Normalize CRLF into \n to avoid double escaping
				if i+1 < len(text) && text[i+1] == '\n' {
					builder.WriteString(`\n`)
					i++
				} else {
					builder.WriteString(`\r`)
				}
			} else {
				builder.WriteByte(ch)
			}
		case '\n':
			if inString {
				builder.WriteString(`\n`)
			} else {
				builder.WriteByte(ch)
			}
		case '\t':
			if inString {
				builder.WriteString(`\t`)
			} else {
				builder.WriteByte(ch)
			}
		default:
			builder.WriteByte(ch)
		}
	}

	return builder.String()
}

// extractJSONObject finds and extracts the JSON object from text more aggressively
func extractJSONObject(text string) string {
	// Find the first { and last } that form a valid JSON object
	firstBrace := strings.Index(text, "{")
	if firstBrace == -1 {
		return text
	}

	// Count braces to find the matching closing brace
	braceCount := 0
	lastValidBrace := -1
	inString := false
	escape := false

	for i := firstBrace; i < len(text); i++ {
		char := text[i]

		// Handle escape sequences
		if escape {
			escape = false
			continue
		}
		if char == '\\' {
			escape = true
			continue
		}

		// Handle strings
		if char == '"' {
			inString = !inString
			continue
		}
		if inString {
			continue
		}

		// Count braces outside of strings
		if char == '{' {
			braceCount++
		} else if char == '}' {
			braceCount--
			if braceCount == 0 {
				lastValidBrace = i
				break
			}
		}
	}

	if lastValidBrace > firstBrace {
		return text[firstBrace : lastValidBrace+1]
	}

	return text
}
