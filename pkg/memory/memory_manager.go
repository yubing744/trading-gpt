package memory

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// MemoryManager handles file-based memory operations
type MemoryManager struct {
	memoryPath string
	maxWords   int
}

// NewMemoryManager creates a new memory manager
func NewMemoryManager(memoryPath string, maxWords int) *MemoryManager {
	return &MemoryManager{
		memoryPath: memoryPath,
		maxWords:   maxWords,
	}
}

// LoadMemory loads memory content from file
func (m *MemoryManager) LoadMemory() (string, error) {
	if _, err := os.Stat(m.memoryPath); os.IsNotExist(err) {
		return "", nil // File doesn't exist, return empty memory
	}

	content, err := os.ReadFile(m.memoryPath)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

// SaveMemory saves memory content to file with word limit enforcement
func (m *MemoryManager) SaveMemory(content string) (string, bool, error) {
	originalWordCount := len(strings.Fields(content))
	truncated := m.truncateToMaxWords(content)

	wasTruncated := originalWordCount > m.maxWords

	// Ensure the directory exists before writing the file
	dir := filepath.Dir(m.memoryPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", false, fmt.Errorf("failed to create memory directory: %w", err)
	}

	err := os.WriteFile(m.memoryPath, []byte(truncated), 0644)
	if err != nil {
		return "", false, fmt.Errorf("failed to write memory file: %w", err)
	}

	return truncated, wasTruncated, nil
}

// truncateToMaxWords truncates content to maximum word limit
func (m *MemoryManager) truncateToMaxWords(content string) string {
	words := strings.Fields(content)
	if len(words) <= m.maxWords {
		return content
	}

	// Keep the latest memory content
	return strings.Join(words[len(words)-m.maxWords:], " ")
}

// GetWordLimitInfo returns word limit information for AI feedback
func (m *MemoryManager) GetWordLimitInfo() string {
	return fmt.Sprintf("Memory word limit: %d words", m.maxWords)
}

// GetMaxWords returns the maximum word limit
func (m *MemoryManager) GetMaxWords() int {
	return m.maxWords
}
