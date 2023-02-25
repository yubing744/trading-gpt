package feishu

import (
	"context"
	"os"

	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	"github.com/sirupsen/logrus"
	"github.com/yubing744/trading-bot/pkg/config"
	"github.com/yubing744/trading-bot/pkg/types"

	chatfeishu "github.com/yubing744/trading-bot/pkg/chat/feishu"
)

var log = logrus.WithField("notify", "feishu")

type FeishuNotifyChannel struct {
	channel *chatfeishu.FeishuChatChannel
}

func NewFeishuNotifyChannel(cfg *config.NotifyFeishuConfig) *FeishuNotifyChannel {
	opts := []lark.ClientOptionFunc{}
	debugMock := os.Getenv("DEBUG")
	if debugMock == "true" {
		opts = append(opts, lark.WithLogReqAtDebug(true))
		opts = append(opts, lark.WithLogLevel(larkcore.LogLevelDebug))
	}
	cli := lark.NewClient(cfg.AppId, cfg.AppSecret, opts...)

	if cfg.TenantKey == "" || cfg.ReceiveIdType == "" || cfg.ReceiveId == "" {
		log.Fatal("TenantKey OR ReceiveIdType OR ReceiveId required!")
	}

	log.WithField("tenant_key", cfg.TenantKey).
		WithField("receive_id_type", cfg.ReceiveIdType).
		WithField("Receive_id", cfg.ReceiveId).
		Info("create feishu notify channel")

	return &FeishuNotifyChannel{
		channel: chatfeishu.NewFeishuChatChannel(cli, cfg.TenantKey, cfg.ReceiveIdType, cfg.ReceiveId),
	}
}

func (ch *FeishuNotifyChannel) GetID() string {
	return ch.channel.GetID()
}

func (ch *FeishuNotifyChannel) Reply(ctx context.Context, msg *types.Message) error {
	return ch.channel.Reply(ctx, msg)
}
