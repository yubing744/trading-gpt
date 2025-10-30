package config

import (
	"github.com/c9s/bbgo/pkg/fixedpoint"
	"github.com/c9s/bbgo/pkg/types"
)

type Config struct {
	Symbol             string           `json:"symbol"`
	Interval           types.Interval   `json:"interval"`
	SubscribeIntervals []types.Interval `json:"subscribe_intervals"`
	Leverage           fixedpoint.Value `json:"leverage"`
	MaxNum             int              `json:"max_num"`

	Notify NotifyConfig `json:"notify"`
	LLM    LLMConfig    `json:"llm"`
	Chat   ChatConfig   `json:"chat"`
	Agent  AgentConfig  `json:"agent"`
	Env    EnvConfig    `json:"env"`

	Strategy                string   `json:"strategy"`
	StrategyAttentionPoints []string `json:"strategy_attention_points"`

	// ReflectionPath specifies the directory path where trade reflections will be stored
	// If not specified, defaults to "memory-bank/reflections/"
	ReflectionPath string `json:"reflection_path"`

	// ReflectionEnabled controls whether trade reflections are generated and saved
	// If not specified, defaults to true
	ReflectionEnabled *bool `json:"reflection_enabled,omitempty"`

	// ReadMemoryEnabled controls whether the system reads from memory bank reflections
	// If not specified, defaults to true
	ReadMemoryEnabled *bool `json:"read_memory_enabled,omitempty"`

	// Memory configuration for file-based memory function
	Memory MemoryConfig `json:"memory"`

	// Commands configuration for next-cycle command persistence
	Commands CommandsConfig `json:"commands"`
}

// MemoryConfig defines configuration for the file-based memory system
type MemoryConfig struct {
	Enabled    bool   `json:"enabled"`     // Whether to enable memory function
	MemoryPath string `json:"memory_path"` // Path to memory file
	MaxWords   int    `json:"max_words"`   // Maximum word limit for memory
}

// CommandsConfig defines configuration for the command persistence system
type CommandsConfig struct {
	Enabled     bool   `json:"enabled"`      // Whether to enable command system
	CommandPath string `json:"command_path"` // Path to command persistence file
}
