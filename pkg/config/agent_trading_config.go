package config

type TradingAgentConfig struct {
	Enabled          bool    `json:"enabled"`
	Name             string  `json:"name"`
	Model            string  `json:"model"`
	Temperature      float32 `json:"temperature"`
	MaxContextLength int     `json:"max_context_length"`
	LLM              string  `json:"llm"`
	Backgroup        string  `json:"backgroup"`
}
