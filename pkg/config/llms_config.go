package config

type OpenAIConfig struct {
	Token   string `json:"token"`
	Model   string `json:"model"`
	BaseURL string `json:"base_url"`
}

type OllamaConfig struct {
	Model     string `json:"model"`
	ServerURL string `json:"server_url"`
	Format    string `json:"format"`
}

type AnthropicConfig struct {
	Token   string `json:"token"`
	Model   string `json:"model"`
	BaseURL string `json:"base_url"`
}

type LLMConfig struct {
	Primary   string           `json:"primary,omitempty"`
	Backup    string           `json:"backup,omitempty"`
	OpenAI    *OpenAIConfig    `json:"openai,omitempty"`
	Ollama    *OllamaConfig    `json:"ollama,omitempty"`
	Anthropic *AnthropicConfig `json:"anthropic,omitempty"`
}
