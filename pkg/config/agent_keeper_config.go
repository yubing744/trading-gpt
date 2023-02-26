package config

type AgentKeeperConfig struct {
	Enabled   bool     `json:"enabled"`
	Leader    string   `json:"leader"`
	Followers []string `json:"followers"`
}
