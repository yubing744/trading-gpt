package config

type AgentOpenAIConfig struct {
	Enabled          bool   `json:"enabled"`
	Name             string `json:"name"`
	Token            string `json:"token"`
	Model            string `json:"model"`
	Backgroup        string `json:"backgroup"`
	MaxContextLength int    `json:"max_context_length"`
}
