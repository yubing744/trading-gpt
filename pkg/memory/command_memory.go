package memory

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// PendingCommand represents a command waiting to be executed
type PendingCommand struct {
	ID          string            `json:"id"`           // Unique command ID
	EntityID    string            `json:"entity_id"`    // Target entity
	CommandName string            `json:"command_name"` // Command/workflow name
	Args        map[string]string `json:"args"`         // Parameters
	Status      string            `json:"status"`       // pending/completed/failed
	RetryCount  int               `json:"retry_count"`  // Current retry attempt
	MaxRetries  int               `json:"max_retries"`  // Max retry limit
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	Error       string            `json:"error,omitempty"` // Error message if failed
}

// CommandStore represents the structure of the commands JSON file
type CommandStore struct {
	Pending   []*PendingCommand `json:"pending"`
	Completed []*PendingCommand `json:"completed"`
	Failed    []*PendingCommand `json:"failed"`
}

// CommandMemory handles file-based command persistence
type CommandMemory struct {
	commandPath string
}

// NewCommandMemory creates a new command memory manager
func NewCommandMemory(commandPath string) *CommandMemory {
	return &CommandMemory{
		commandPath: commandPath,
	}
}

// LoadPendingCommands loads pending commands from file
func (cm *CommandMemory) LoadPendingCommands() ([]*PendingCommand, error) {
	store, err := cm.loadStore()
	if err != nil {
		return nil, err
	}

	// Return only pending commands and failed commands that can be retried
	result := make([]*PendingCommand, 0)
	for _, cmd := range store.Pending {
		if cmd.Status == "pending" || (cmd.Status == "failed" && cmd.RetryCount < cmd.MaxRetries) {
			result = append(result, cmd)
		}
	}

	return result, nil
}

// SaveCommands saves commands to file, organizing by status
func (cm *CommandMemory) SaveCommands(commands []*PendingCommand) error {
	// Load existing store
	store, err := cm.loadStore()
	if err != nil {
		// If file doesn't exist or is corrupted, start fresh
		store = &CommandStore{
			Pending:   make([]*PendingCommand, 0),
			Completed: make([]*PendingCommand, 0),
			Failed:    make([]*PendingCommand, 0),
		}
	}

	// Organize commands by status
	for _, cmd := range commands {
		switch cmd.Status {
		case "pending":
			// Check if command already exists in pending
			found := false
			for i, existing := range store.Pending {
				if existing.ID == cmd.ID {
					store.Pending[i] = cmd
					found = true
					break
				}
			}
			if !found {
				store.Pending = append(store.Pending, cmd)
			}
		case "completed":
			// Move to completed, remove from pending
			cm.removeFromPending(store, cmd.ID)
			store.Completed = append(store.Completed, cmd)
		case "failed":
			if cmd.RetryCount >= cmd.MaxRetries {
				// Permanently failed, move to failed list
				cm.removeFromPending(store, cmd.ID)
				store.Failed = append(store.Failed, cmd)
			} else {
				// Still can retry, keep in pending with updated status
				for i, existing := range store.Pending {
					if existing.ID == cmd.ID {
						store.Pending[i] = cmd
						break
					}
				}
			}
		}
	}

	return cm.saveStore(store)
}

// ArchiveCompletedCommands removes old completed commands to keep file size manageable
func (cm *CommandMemory) ArchiveCompletedCommands() error {
	store, err := cm.loadStore()
	if err != nil {
		return err
	}

	// Keep only recent completed commands (e.g., last 50)
	maxCompleted := 50
	if len(store.Completed) > maxCompleted {
		store.Completed = store.Completed[len(store.Completed)-maxCompleted:]
	}

	// Keep only recent failed commands (e.g., last 50)
	maxFailed := 50
	if len(store.Failed) > maxFailed {
		store.Failed = store.Failed[len(store.Failed)-maxFailed:]
	}

	return cm.saveStore(store)
}

// loadStore loads the command store from file
func (cm *CommandMemory) loadStore() (*CommandStore, error) {
	if _, err := os.Stat(cm.commandPath); os.IsNotExist(err) {
		// File doesn't exist, return empty store
		return &CommandStore{
			Pending:   make([]*PendingCommand, 0),
			Completed: make([]*PendingCommand, 0),
			Failed:    make([]*PendingCommand, 0),
		}, nil
	}

	content, err := os.ReadFile(cm.commandPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read command file: %w", err)
	}

	var store CommandStore
	if err := json.Unmarshal(content, &store); err != nil {
		return nil, fmt.Errorf("failed to parse command file: %w", err)
	}

	return &store, nil
}

// saveStore saves the command store to file using atomic write
func (cm *CommandMemory) saveStore(store *CommandStore) error {
	// Ensure directory exists
	dir := filepath.Dir(cm.commandPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create command directory: %w", err)
	}

	// Marshal to JSON with indentation for readability
	data, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal command store: %w", err)
	}

	// Atomic write: write to temp file, then rename
	tempPath := cm.commandPath + ".tmp"
	if err := os.WriteFile(tempPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write temp command file: %w", err)
	}

	if err := os.Rename(tempPath, cm.commandPath); err != nil {
		os.Remove(tempPath) // Clean up temp file on failure
		return fmt.Errorf("failed to rename temp command file: %w", err)
	}

	return nil
}

// removeFromPending removes a command from the pending list
func (cm *CommandMemory) removeFromPending(store *CommandStore, cmdID string) {
	for i, cmd := range store.Pending {
		if cmd.ID == cmdID {
			store.Pending = append(store.Pending[:i], store.Pending[i+1:]...)
			break
		}
	}
}
