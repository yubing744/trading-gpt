package utils

import (
	"testing"
)

func TestExtractThinking(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectedHas     bool
		expectedContent string
	}{
		{
			name:            "empty string",
			input:           "",
			expectedHas:     false,
			expectedContent: "",
		},
		{
			name:            "no thinking tags",
			input:           "This is just regular content",
			expectedHas:     false,
			expectedContent: "",
		},
		{
			name:            "only start tag",
			input:           "<thinking>Some thinking content",
			expectedHas:     false,
			expectedContent: "",
		},
		{
			name:            "only end tag",
			input:           "Some content</thinking>",
			expectedHas:     false,
			expectedContent: "",
		},
		{
			name:            "end tag before start tag",
			input:           "Content</thinking>more content<thinking>",
			expectedHas:     false,
			expectedContent: "",
		},
		{
			name:            "valid thinking tags",
			input:           "<thinking>This is thinking content</thinking>",
			expectedHas:     true,
			expectedContent: "This is thinking content",
		},
		{
			name:            "thinking with surrounding content",
			input:           "Before content <thinking>thinking content</thinking> After content",
			expectedHas:     true,
			expectedContent: "thinking content",
		},
		{
			name:            "thinking with whitespace",
			input:           "<thinking>\n  This is thinking content  \n</thinking>",
			expectedHas:     true,
			expectedContent: "This is thinking content",
		},
		{
			name:            "empty thinking tags",
			input:           "<thinking></thinking>",
			expectedHas:     true,
			expectedContent: "",
		},
		{
			name:            "thinking with only whitespace",
			input:           "<thinking>   \n  \t  </thinking>",
			expectedHas:     true,
			expectedContent: "",
		},
		{
			name:            "multiple thinking blocks - first one is extracted",
			input:           "<thinking>first</thinking> content <thinking>second</thinking>",
			expectedHas:     true,
			expectedContent: "first",
		},
		{
			name:            "thinking with multiline content",
			input:           "<thinking>Line 1\nLine 2\nLine 3</thinking>",
			expectedHas:     true,
			expectedContent: "Line 1\nLine 2\nLine 3",
		},
		{
			name:            "thinking with special characters",
			input:           "<thinking>Special chars: @#$%^&*(){}[]</thinking>",
			expectedHas:     true,
			expectedContent: "Special chars: @#$%^&*(){}[]",
		},
		{
			name:            "case sensitive tags",
			input:           "<THINKING>upper case</THINKING>",
			expectedHas:     false,
			expectedContent: "",
		},
		{
			name:            "nested angle brackets in content",
			input:           "<thinking>Content with <tags> inside</thinking>",
			expectedHas:     true,
			expectedContent: "Content with <tags> inside",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasThinking, content := ExtractThinking(tt.input)
			
			if hasThinking != tt.expectedHas {
				t.Errorf("ExtractThinking() hasThinking = %v, want %v", hasThinking, tt.expectedHas)
			}
			
			if content != tt.expectedContent {
				t.Errorf("ExtractThinking() content = %q, want %q", content, tt.expectedContent)
			}
		})
	}
}

func TestExtractThinkingFull(t *testing.T) {
	tests := []struct {
		name              string
		input             string
		expectedHas       bool
		expectedThinking  string
		expectedRemaining string
	}{
		{
			name:              "empty string",
			input:             "",
			expectedHas:       false,
			expectedThinking:  "",
			expectedRemaining: "",
		},
		{
			name:              "no thinking tags",
			input:             "This is just regular content",
			expectedHas:       false,
			expectedThinking:  "",
			expectedRemaining: "This is just regular content",
		},
		{
			name:              "only start tag",
			input:             "<thinking>Some thinking content",
			expectedHas:       false,
			expectedThinking:  "",
			expectedRemaining: "<thinking>Some thinking content",
		},
		{
			name:              "only end tag",
			input:             "Some content</thinking>",
			expectedHas:       false,
			expectedThinking:  "",
			expectedRemaining: "Some content</thinking>",
		},
		{
			name:              "end tag before start tag",
			input:             "Content</thinking>more content<thinking>",
			expectedHas:       false,
			expectedThinking:  "",
			expectedRemaining: "Content</thinking>more content<thinking>",
		},
		{
			name:              "valid thinking tags only",
			input:             "<thinking>This is thinking content</thinking>",
			expectedHas:       true,
			expectedThinking:  "This is thinking content",
			expectedRemaining: "",
		},
		{
			name:              "thinking with content before",
			input:             "Before content <thinking>thinking content</thinking>",
			expectedHas:       true,
			expectedThinking:  "thinking content",
			expectedRemaining: "Before content",
		},
		{
			name:              "thinking with content after",
			input:             "<thinking>thinking content</thinking> After content",
			expectedHas:       true,
			expectedThinking:  "thinking content",
			expectedRemaining: "After content",
		},
		{
			name:              "thinking with content before and after",
			input:             "Before content <thinking>thinking content</thinking> After content",
			expectedHas:       true,
			expectedThinking:  "thinking content",
			expectedRemaining: "Before content After content",
		},
		{
			name:              "thinking with whitespace around content",
			input:             "  Before  <thinking>  thinking  </thinking>  After  ",
			expectedHas:       true,
			expectedThinking:  "thinking",
			expectedRemaining: "Before After",
		},
		{
			name:              "empty thinking tags",
			input:             "Before <thinking></thinking> After",
			expectedHas:       true,
			expectedThinking:  "",
			expectedRemaining: "Before After",
		},
		{
			name:              "thinking with only whitespace",
			input:             "Before <thinking>   \n  \t  </thinking> After",
			expectedHas:       true,
			expectedThinking:  "",
			expectedRemaining: "Before After",
		},
		{
			name:              "multiple thinking blocks - first one processed",
			input:             "<thinking>first</thinking> middle <thinking>second</thinking> end",
			expectedHas:       true,
			expectedThinking:  "first",
			expectedRemaining: "middle <thinking>second</thinking> end",
		},
		{
			name:              "multiline content with thinking",
			input:             "Line 1\n<thinking>thinking\ncontent</thinking>\nLine 2",
			expectedHas:       true,
			expectedThinking:  "thinking\ncontent",
			expectedRemaining: "Line 1 Line 2",
		},
		{
			name:              "thinking at start with no before content",
			input:             "<thinking>thinking content</thinking> After content",
			expectedHas:       true,
			expectedThinking:  "thinking content",
			expectedRemaining: "After content",
		},
		{
			name:              "thinking at end with no after content",
			input:             "Before content <thinking>thinking content</thinking>",
			expectedHas:       true,
			expectedThinking:  "thinking content",
			expectedRemaining: "Before content",
		},
		{
			name:              "thinking with empty before content and whitespace",
			input:             "   <thinking>thinking</thinking> after",
			expectedHas:       true,
			expectedThinking:  "thinking",
			expectedRemaining: "after",
		},
		{
			name:              "thinking with empty after content and whitespace",
			input:             "before <thinking>thinking</thinking>   ",
			expectedHas:       true,
			expectedThinking:  "thinking",
			expectedRemaining: "before",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasThinking, thinking, remaining := ExtractThinkingFull(tt.input)
			
			if hasThinking != tt.expectedHas {
				t.Errorf("ExtractThinkingFull() hasThinking = %v, want %v", hasThinking, tt.expectedHas)
			}
			
			if thinking != tt.expectedThinking {
				t.Errorf("ExtractThinkingFull() thinking = %q, want %q", thinking, tt.expectedThinking)
			}
			
			if remaining != tt.expectedRemaining {
				t.Errorf("ExtractThinkingFull() remaining = %q, want %q", remaining, tt.expectedRemaining)
			}
		})
	}
}

func TestExtractThinkingSimple(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectedHas     bool
		expectedContent string
	}{
		{
			name:            "valid thinking",
			input:           "<thinking>content</thinking>",
			expectedHas:     true,
			expectedContent: "content",
		},
		{
			name:            "no thinking",
			input:           "regular content",
			expectedHas:     false,
			expectedContent: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasThinking, content := ExtractThinkingSimple(tt.input)
			
			if hasThinking != tt.expectedHas {
				t.Errorf("ExtractThinkingSimple() hasThinking = %v, want %v", hasThinking, tt.expectedHas)
			}
			
			if content != tt.expectedContent {
				t.Errorf("ExtractThinkingSimple() content = %q, want %q", content, tt.expectedContent)
			}
		})
	}
}

func TestRemoveThinkingTags(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no thinking tags",
			input:    "regular content",
			expected: "regular content",
		},
		{
			name:     "with thinking tags",
			input:    "before <thinking>thinking</thinking> after",
			expected: "before after",
		},
		{
			name:     "only thinking tags",
			input:    "<thinking>only thinking</thinking>",
			expected: "",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RemoveThinkingTags(tt.input)
			if result != tt.expected {
				t.Errorf("RemoveThinkingTags() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestIsThinkingResponse(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "has both tags",
			input:    "<thinking>content</thinking>",
			expected: true,
		},
		{
			name:     "has both tags with other content",
			input:    "before <thinking>content</thinking> after",
			expected: true,
		},
		{
			name:     "only start tag",
			input:    "<thinking>content",
			expected: false,
		},
		{
			name:     "only end tag",
			input:    "content</thinking>",
			expected: false,
		},
		{
			name:     "no tags",
			input:    "regular content",
			expected: false,
		},
		{
			name:     "empty string",
			input:    "",
			expected: false,
		},
		{
			name:     "tags in wrong order",
			input:    "</thinking>content<thinking>",
			expected: true, // function only checks for presence, not order
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsThinkingResponse(tt.input)
			if result != tt.expected {
				t.Errorf("IsThinkingResponse() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestFormatThinking(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "simple content",
			input:    "thinking content",
			expected: "<thinking>\nthinking content\n</thinking>",
		},
		{
			name:     "content with whitespace",
			input:    "  thinking content  ",
			expected: "<thinking>\nthinking content\n</thinking>",
		},
		{
			name:     "multiline content",
			input:    "line 1\nline 2",
			expected: "<thinking>\nline 1\nline 2\n</thinking>",
		},
		{
			name:     "only whitespace",
			input:    "   \n  \t  ",
			expected: "<thinking>\n\n</thinking>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatThinking(tt.input)
			if result != tt.expected {
				t.Errorf("FormatThinking() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestCleanThinkingText(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "simple text",
			input:    "simple text",
			expected: "simple text",
		},
		{
			name:     "text with leading/trailing whitespace",
			input:    "  text with spaces  ",
			expected: "text with spaces",
		},
		{
			name:     "text with CRLF line endings",
			input:    "line1\r\nline2\r\nline3",
			expected: "line1\nline2\nline3",
		},
		{
			name:     "text with excessive newlines",
			input:    "line1\n\n\n\nline2\n\n\n\n\nline3",
			expected: "line1\n\nline2\n\nline3",
		},
		{
			name:     "complex text with all issues",
			input:    "  line1\r\n\r\n\r\n\r\nline2\n\n\n\nline3  ",
			expected: "line1\n\nline2\n\nline3",
		},
		{
			name:     "only whitespace",
			input:    "   \t  \n  ",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CleanThinkingText(tt.input)
			if result != tt.expected {
				t.Errorf("CleanThinkingText() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestSplitThinkingAndContent(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectedThinking string
		expectedContent  string
	}{
		{
			name:            "no thinking",
			input:           "regular content",
			expectedThinking: "",
			expectedContent:  "regular content",
		},
		{
			name:            "with thinking",
			input:           "before <thinking>thinking</thinking> after",
			expectedThinking: "thinking",
			expectedContent:  "before after",
		},
		{
			name:            "only thinking",
			input:           "<thinking>only thinking</thinking>",
			expectedThinking: "only thinking",
			expectedContent:  "",
		},
		{
			name:            "empty string",
			input:           "",
			expectedThinking: "",
			expectedContent:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			thinking, content := SplitThinkingAndContent(tt.input)
			
			if thinking != tt.expectedThinking {
				t.Errorf("SplitThinkingAndContent() thinking = %q, want %q", thinking, tt.expectedThinking)
			}
			
			if content != tt.expectedContent {
				t.Errorf("SplitThinkingAndContent() content = %q, want %q", content, tt.expectedContent)
			}
		})
	}
}