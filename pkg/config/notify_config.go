package config

type NotifyConfig struct {
	Feishu     *NotifyFeishuConfig     `json:"feishu"`
	FeishuHook *NotifyFeishuHookConfig `json:"feishu_hook"`
}
