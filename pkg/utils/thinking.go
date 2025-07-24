package utils

import (
	"regexp"
	"strings"
)

const (
	// ThinkingStartMarker marks the beginning of thinking content
	ThinkingStartMarker = "<thinking>"
	// ThinkingEndMarker marks the end of thinking content
	ThinkingEndMarker = "</thinking>"
)

// ExtractThinking extracts thinking content from Claude's response
// Returns (hasThinking, thinkingText)
// Deprecated: Use ExtractThinkingFull for full functionality
func ExtractThinking(text string) (bool, string) {
	hasThinking, thinkingText, _ := ExtractThinkingFull(text)
	return hasThinking, thinkingText
}

// ExtractThinkingFull extracts thinking content from Claude's response
// Returns (hasThinking, thinkingText, remainingContent)
func ExtractThinkingFull(text string) (bool, string, string) {
	if text == "" {
		return false, "", ""
	}

	startIdx := strings.Index(text, ThinkingStartMarker)
	endIdx := strings.Index(text, ThinkingEndMarker)

	if startIdx == -1 || endIdx == -1 || startIdx >= endIdx {
		return false, "", text
	}

	// Extract thinking content
	startPos := startIdx + len(ThinkingStartMarker)
	thinkingText := strings.TrimSpace(text[startPos:endIdx])

	// Extract remaining content
	beforeThinking := strings.TrimSpace(text[:startIdx])
	afterThinking := strings.TrimSpace(text[endIdx+len(ThinkingEndMarker):])

	// Combine remaining content
	var remainingContent string
	if beforeThinking == "" {
		remainingContent = afterThinking
	} else if afterThinking == "" {
		remainingContent = beforeThinking
	} else {
		remainingContent = beforeThinking + " " + afterThinking
	}

	return true, thinkingText, remainingContent
}

// ExtractThinkingSimple is a simpler version that just returns thinking text
// Returns (hasThinking, thinkingText)
func ExtractThinkingSimple(text string) (bool, string) {
	hasThinking, thinkingText, _ := ExtractThinkingFull(text)
	return hasThinking, thinkingText
}

// RemoveThinkingTags removes thinking tags from the text
func RemoveThinkingTags(text string) string {
	_, _, remaining := ExtractThinkingFull(text)
	return remaining
}

// IsThinkingResponse checks if the response contains thinking content
func IsThinkingResponse(text string) bool {
	return strings.Contains(text, ThinkingStartMarker) && strings.Contains(text, ThinkingEndMarker)
}

// FormatThinking formats thinking content with proper markers
func FormatThinking(thinkingText string) string {
	if thinkingText == "" {
		return ""
	}
	return ThinkingStartMarker + "\n" + strings.TrimSpace(thinkingText) + "\n" + ThinkingEndMarker
}

// CleanThinkingText removes extra whitespace and normalizes thinking text
func CleanThinkingText(text string) string {
	if text == "" {
		return ""
	}

	// Remove leading/trailing whitespace
	text = strings.TrimSpace(text)

	// Normalize line endings
	text = strings.ReplaceAll(text, "\r\n", "\n")

	// Remove excessive newlines
	re := regexp.MustCompile(`\n{3,}`)
	text = re.ReplaceAllString(text, "\n\n")

	return text
}

// SplitThinkingAndContent splits the response into thinking and content parts
func SplitThinkingAndContent(text string) (thinking string, content string) {
	_, thinking, content = ExtractThinkingFull(text)
	return thinking, content
}
