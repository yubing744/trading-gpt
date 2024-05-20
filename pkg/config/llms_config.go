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

type GoogleAIConfig struct {
	APIKey string `json:"api_key"`
	Model  string `json:"model"`
}

type LLMConfig struct {
	Primary   string           `json:"primary,omitempty"`
	Secondly  string           `json:"secondly,omitempty"`
	OpenAI    *OpenAIConfig    `json:"openai,omitempty"`
	Ollama    *OllamaConfig    `json:"ollama,omitempty"`
	Anthropic *AnthropicConfig `json:"anthropic,omitempty"`
	GoogleAI  *GoogleAIConfig  `json:"googleai,omitempty"`
}
