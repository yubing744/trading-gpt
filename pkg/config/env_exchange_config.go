package config

type EnvExchangeConfig struct {
	WindowSize    int      `json:"window_size"`
	IncludeEvents []string `json:"include_events"`
}
