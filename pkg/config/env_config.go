package config

type EnvConfig struct {
	ExchangeConfig *EnvExchangeConfig      `json:"exchange"`
	FNG            *FNGConfig              `json:"fng"`
	Coze           *CozeEntityConfig       `json:"coze"`
	TwitterAPI     *TwitterAPIEntityConfig `json:"twitterapi"`
	IncludeEvents  []string                `json:"include_events"`
}
