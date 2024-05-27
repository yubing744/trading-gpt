package config

type EnvConfig struct {
	ExchangeConfig *EnvExchangeConfig `json:"exchange"`
	FNG            *FNGConfig         `json:"fng"`
	Coze           *CozeEntityConfig  `json:"coze"`
	IncludeEvents  []string           `json:"include_events"`
}
