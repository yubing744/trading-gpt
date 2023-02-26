package config

type EnvConfig struct {
	ExchangeConfig *EnvExchangeConfig `json:"exchange"`
	IncludeEvents  []string           `json:"include_events"`
}
