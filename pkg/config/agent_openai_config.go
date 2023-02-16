package config

type AgentOpenAIConfig struct {
	Token            string `json:"token"`
	MaxContextLength int    `json:"max_context_length"`
}
