package config

type LLMConfig struct {
	OpenAI *LLMOpenAIConfig `json:"openai"`
}
