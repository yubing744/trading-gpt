package memory

import (
	"os"
	"testing"
)

func TestMemoryManager(t *testing.T) {
	// Create a temporary file for testing
	tmpFile := "test_memory.md"
	defer os.Remove(tmpFile) // Clean up after test

	// Create memory manager with small word limit for testing
	mm := NewMemoryManager(tmpFile, 10)

	// Test loading non-existent file
	content, err := mm.LoadMemory()
	if err != nil {
		t.Fatalf("Failed to load memory from non-existent file: %v", err)
	}
	if content != "" {
		t.Errorf("Expected empty content for non-existent file, got: %s", content)
	}

	// Test saving memory
	testContent := "This is a test memory content with more than ten words to test truncation functionality"
	savedContent, wasTruncated, err := mm.SaveMemory(testContent)
	if err != nil {
		t.Fatalf("Failed to save memory: %v", err)
	}
	if !wasTruncated {
		t.Error("Expected content to be truncated, but it wasn't")
	}

	// Test loading saved memory
	loadedContent, err := mm.LoadMemory()
	if err != nil {
		t.Fatalf("Failed to load saved memory: %v", err)
	}
	if loadedContent != savedContent {
		t.Errorf("Loaded content doesn't match saved content. Expected: %s, Got: %s", savedContent, loadedContent)
	}

	// Test word limit info
	limitInfo := mm.GetWordLimitInfo()
	expectedLimitInfo := "Memory word limit: 10 words"
	if limitInfo != expectedLimitInfo {
		t.Errorf("Expected limit info: %s, Got: %s", expectedLimitInfo, limitInfo)
	}

	// Test max words
	maxWords := mm.GetMaxWords()
	if maxWords != 10 {
		t.Errorf("Expected max words: 10, Got: %d", maxWords)
	}
}

func TestMemoryManagerWithNonExistentDirectory(t *testing.T) {
	// Create a memory manager with a path in a non-existent directory
	tmpFile := "test_dir/subdir/memory.md"
	defer os.RemoveAll("test_dir") // Clean up after test

	mm := NewMemoryManager(tmpFile, 10)

	// Test saving memory to non-existent directory
	testContent := "This is a test memory content"
	savedContent, wasTruncated, err := mm.SaveMemory(testContent)
	if err != nil {
		t.Fatalf("Failed to save memory to non-existent directory: %v", err)
	}
	if wasTruncated {
		t.Error("Expected content not to be truncated")
	}
	if savedContent != testContent {
		t.Errorf("Expected saved content to match input, got: %s", savedContent)
	}

	// Verify the file was created
	if _, err := os.Stat(tmpFile); os.IsNotExist(err) {
		t.Error("Memory file was not created")
	}

	// Test loading the saved memory
	loadedContent, err := mm.LoadMemory()
	if err != nil {
		t.Fatalf("Failed to load memory: %v", err)
	}
	if loadedContent != testContent {
		t.Errorf("Expected loaded content to match saved content. Expected: %s, Got: %s", testContent, loadedContent)
	}
}
