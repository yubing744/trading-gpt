package config

import (
	"github.com/c9s/bbgo/pkg/fixedpoint"
	"github.com/c9s/bbgo/pkg/types"
)

type Config struct {
	Symbol        string           `json:"symbol"`
	Interval      types.Interval   `json:"interval"`
	Leverage      fixedpoint.Value `json:"leverage"`
	MaxWindowSize int              `json:"max_window_size"`

	Notify NotifyConfig `json:"notify"`
	Chat   ChatConfig   `json:"chat"`
	LLM    LLMConfig    `json:"llm"`
	Env    EnvConfig    `json:"env"`

	Strategy string `json:"strategy"`
}
