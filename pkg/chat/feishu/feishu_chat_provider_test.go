package feishu

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yubing744/trading-bot/pkg/config"
)

func TestNewFeishuChatProvider(t *testing.T) {
	cfg := &config.ChatFeishuConfig{
		AppId:             "xxxxx",
		AppSecret:         "xxxx",
		EventEncryptKey:   "xxxxx",
		VerificationToken: "xxxx",
	}
	feishu := NewFeishuChatProvider(cfg)
	assert.NotNil(t, feishu)
}
