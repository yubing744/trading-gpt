package config

type KeeperAgentConfig struct {
	Enabled   bool     `json:"enabled"`
	Leader    string   `json:"leader"`
	Followers []string `json:"followers"`
}
