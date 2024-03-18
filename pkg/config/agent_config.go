package config

type AgentConfig struct {
	Trading TradingAgentConfig `json:"trading"`
	Keeper  KeeperAgentConfig  `json:"keeper"`
}
