package config

type AgentOpenAIConfig struct {
	Enabled          bool   `json:"enabled"`
	Name             string `json:"name"`
	Token            string `json:"token"`
	MaxContextLength int    `json:"max_context_length"`
}
