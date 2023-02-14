package config

type Config struct {
	Chat  ChatConfig  `json:"chat"`
	Agent AgentConfig `json:"agent"`
}
