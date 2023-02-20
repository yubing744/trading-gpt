package config

type AgentChatGPTConfig struct {
	Enabled          bool   `json:"enabled"`
	Name             string `json:"name"`
	Email            string `json:"email"`
	Password         string `json:"password"`
	MaxContextLength int    `json:"max_context_length"`
}
