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
}
