package config

type ChatFeishuConfig struct {
	AppId             string `json:"app_id"`
	AppSecret         string `json:"app_secret"`
	VerificationToken string `json:"verification_token"`
	EventEncryptKey   string `json:"event_encrypt_key"`
	ServerPort        int    `json:"server_port"`
}
