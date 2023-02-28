package config

type NotifyFeishuHookConfig struct {
	Enabled bool   `json:"enabled"`
	URL     string `json:"url"`
}
