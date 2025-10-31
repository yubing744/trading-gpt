package memory

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCommandMemory_SaveAndLoad(t *testing.T) {
	// Create temp directory for test
	tempDir := t.TempDir()
	commandPath := filepath.Join(tempDir, "commands.json")

	cm := NewCommandMemory(commandPath)

	// Create test commands
	commands := []*PendingCommand{
		{
			ID:          "cmd1",
			EntityID:    "coze",
			CommandName: "workflow_test",
			Args:        map[string]string{"param1": "value1"},
			Status:      "pending",
			RetryCount:  0,
			MaxRetries:  1,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          "cmd2",
			EntityID:    "exchange",
			CommandName: "get_balance",
			Args:        map[string]string{},
			Status:      "pending",
			RetryCount:  0,
			MaxRetries:  1,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	// Save commands
	err := cm.SaveCommands(commands)
	if err != nil {
		t.Fatalf("Failed to save commands: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(commandPath); os.IsNotExist(err) {
		t.Fatal("Command file was not created")
	}

	// Load commands
	loadedCommands, err := cm.LoadPendingCommands()
	if err != nil {
		t.Fatalf("Failed to load commands: %v", err)
	}

	// Verify loaded commands
	if len(loadedCommands) != 2 {
		t.Fatalf("Expected 2 commands, got %d", len(loadedCommands))
	}

	if loadedCommands[0].ID != "cmd1" {
		t.Errorf("Expected first command ID to be 'cmd1', got '%s'", loadedCommands[0].ID)
	}

	if loadedCommands[1].EntityID != "exchange" {
		t.Errorf("Expected second command entity to be 'exchange', got '%s'", loadedCommands[1].EntityID)
	}
}

func TestCommandMemory_StatusTransitions(t *testing.T) {
	tempDir := t.TempDir()
	commandPath := filepath.Join(tempDir, "commands.json")
	cm := NewCommandMemory(commandPath)

	// Create a pending command
	cmd := &PendingCommand{
		ID:          "cmd1",
		EntityID:    "coze",
		CommandName: "test",
		Args:        map[string]string{},
		Status:      "pending",
		RetryCount:  0,
		MaxRetries:  1,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Save as pending
	err := cm.SaveCommands([]*PendingCommand{cmd})
	if err != nil {
		t.Fatalf("Failed to save pending command: %v", err)
	}

	// Mark as completed
	cmd.Status = "completed"
	cmd.UpdatedAt = time.Now()
	err = cm.SaveCommands([]*PendingCommand{cmd})
	if err != nil {
		t.Fatalf("Failed to save completed command: %v", err)
	}

	// Load pending commands - should be empty
	pending, err := cm.LoadPendingCommands()
	if err != nil {
		t.Fatalf("Failed to load pending commands: %v", err)
	}

	if len(pending) != 0 {
		t.Errorf("Expected 0 pending commands after completion, got %d", len(pending))
	}

	// Load store and verify completed list
	store, err := cm.loadStore()
	if err != nil {
		t.Fatalf("Failed to load store: %v", err)
	}

	if len(store.Completed) != 1 {
		t.Errorf("Expected 1 completed command, got %d", len(store.Completed))
	}
}

func TestCommandMemory_RetryLogic(t *testing.T) {
	tempDir := t.TempDir()
	commandPath := filepath.Join(tempDir, "commands.json")
	cm := NewCommandMemory(commandPath)

	// Create a command that will fail
	cmd := &PendingCommand{
		ID:          "cmd1",
		EntityID:    "coze",
		CommandName: "test",
		Args:        map[string]string{},
		Status:      "pending",
		RetryCount:  0,
		MaxRetries:  2,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Save initial command
	err := cm.SaveCommands([]*PendingCommand{cmd})
	if err != nil {
		t.Fatalf("Failed to save command: %v", err)
	}

	// Simulate first failure
	cmd.Status = "failed"
	cmd.RetryCount = 1
	cmd.Error = "test error"
	cmd.UpdatedAt = time.Now()
	err = cm.SaveCommands([]*PendingCommand{cmd})
	if err != nil {
		t.Fatalf("Failed to save failed command: %v", err)
	}

	// Should still be available for retry
	pending, err := cm.LoadPendingCommands()
	if err != nil {
		t.Fatalf("Failed to load pending commands: %v", err)
	}

	if len(pending) != 1 {
		t.Fatalf("Expected 1 pending command for retry, got %d", len(pending))
	}

	// Simulate second failure (exceeds max retries)
	cmd.RetryCount = 2
	cmd.UpdatedAt = time.Now()
	err = cm.SaveCommands([]*PendingCommand{cmd})
	if err != nil {
		t.Fatalf("Failed to save permanently failed command: %v", err)
	}

	// Should be moved to failed list
	pending, err = cm.LoadPendingCommands()
	if err != nil {
		t.Fatalf("Failed to load pending commands: %v", err)
	}

	if len(pending) != 0 {
		t.Errorf("Expected 0 pending commands after max retries, got %d", len(pending))
	}

	// Verify in failed list
	store, err := cm.loadStore()
	if err != nil {
		t.Fatalf("Failed to load store: %v", err)
	}

	if len(store.Failed) != 1 {
		t.Errorf("Expected 1 permanently failed command, got %d", len(store.Failed))
	}
}

func TestCommandMemory_ArchiveOldCommands(t *testing.T) {
	tempDir := t.TempDir()
	commandPath := filepath.Join(tempDir, "commands.json")
	cm := NewCommandMemory(commandPath)

	// Create many completed commands
	commands := make([]*PendingCommand, 60)
	for i := 0; i < 60; i++ {
		commands[i] = &PendingCommand{
			ID:          string(rune(i)),
			EntityID:    "test",
			CommandName: "test",
			Args:        map[string]string{},
			Status:      "completed",
			RetryCount:  0,
			MaxRetries:  1,
			CreatedAt:   time.Now().Add(-time.Duration(i) * time.Hour),
			UpdatedAt:   time.Now(),
		}
	}

	// Save all commands
	err := cm.SaveCommands(commands)
	if err != nil {
		t.Fatalf("Failed to save commands: %v", err)
	}

	// Archive old commands
	err = cm.ArchiveCompletedCommands()
	if err != nil {
		t.Fatalf("Failed to archive commands: %v", err)
	}

	// Load store and verify only 50 remain
	store, err := cm.loadStore()
	if err != nil {
		t.Fatalf("Failed to load store: %v", err)
	}

	if len(store.Completed) > 50 {
		t.Errorf("Expected at most 50 completed commands after archiving, got %d", len(store.Completed))
	}
}

func TestCommandMemory_EmptyFile(t *testing.T) {
	tempDir := t.TempDir()
	commandPath := filepath.Join(tempDir, "nonexistent.json")
	cm := NewCommandMemory(commandPath)

	// Load from non-existent file should return empty list, no error
	commands, err := cm.LoadPendingCommands()
	if err != nil {
		t.Fatalf("Expected no error for non-existent file, got: %v", err)
	}

	if len(commands) != 0 {
		t.Errorf("Expected 0 commands from non-existent file, got %d", len(commands))
	}
}

func TestCommandMemory_AtomicWrite(t *testing.T) {
	tempDir := t.TempDir()
	commandPath := filepath.Join(tempDir, "commands.json")
	cm := NewCommandMemory(commandPath)

	cmd := &PendingCommand{
		ID:          "cmd1",
		EntityID:    "test",
		CommandName: "test",
		Args:        map[string]string{},
		Status:      "pending",
		RetryCount:  0,
		MaxRetries:  1,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Save command
	err := cm.SaveCommands([]*PendingCommand{cmd})
	if err != nil {
		t.Fatalf("Failed to save command: %v", err)
	}

	// Verify temp file doesn't exist
	tempPath := commandPath + ".tmp"
	if _, err := os.Stat(tempPath); !os.IsNotExist(err) {
		t.Error("Temp file should not exist after atomic write")
	}

	// Verify actual file exists
	if _, err := os.Stat(commandPath); os.IsNotExist(err) {
		t.Error("Command file should exist after atomic write")
	}
}
