package config

type AgentConfig struct {
	OpenAI  AgentOpenAIConfig  `json:"openai"`
	ChatGPT AgentChatGPTConfig `json:"chatgpt"`
	Keeper  AgentKeeperConfig  `json:"keeper"`
}
