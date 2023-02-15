package config

import "github.com/c9s/bbgo/pkg/types"

type Config struct {
	Symbol   string         `json:"symbol"`
	Interval types.Interval `json:"interval"`

	Chat  ChatConfig  `json:"chat"`
	Agent AgentConfig `json:"agent"`
	Env   EnvConfig   `json:"env"`
}
