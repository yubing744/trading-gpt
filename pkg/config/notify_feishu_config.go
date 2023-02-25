package config

type NotifyFeishuConfig struct {
	Enabled       bool   `json:"enabled"`
	AppId         string `json:"app_id"`
	AppSecret     string `json:"app_secret"`
	TenantKey     string `json:"tenant_key"`
	ReceiveIdType string `json:"receive_id_type"`
	ReceiveId     string `json:"receive_id"`
}
