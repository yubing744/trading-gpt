package config

type LLMOpenAIConfig struct {
	Enabled bool   `json:"enabled"`
	Token   string `json:"token"`
	Model   string `json:"model"`
}
