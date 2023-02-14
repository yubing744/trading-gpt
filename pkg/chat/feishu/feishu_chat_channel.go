package feishu

import (
	"context"

	"github.com/pkg/errors"
	"github.com/yubing744/trading-bot/pkg/types"

	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

type FeishuChatChannel struct {
	client    *lark.Client
	tenantKey string
	callbacks []types.MessageCallback
}

func NewFeishuChatChannel(client *lark.Client, tenantKey string) *FeishuChatChannel {
	return &FeishuChatChannel{
		client:    client,
		tenantKey: tenantKey,
	}
}

func (ch *FeishuChatChannel) toMessage(event *larkim.P2MessageReceiveV1) *types.Message {
	return &types.Message{
		ID: event.EventReq.RequestId(),
	}
}

func (ch *FeishuChatChannel) handleEvent(event *larkim.P2MessageReceiveV1) {
	msg := ch.toMessage(event)
	for _, cb := range ch.callbacks {
		cb(msg)
	}
}

func (ch *FeishuChatChannel) OnMessage(cb types.MessageCallback) {
	ch.callbacks = append(ch.callbacks, cb)
}

func (ch *FeishuChatChannel) Reply(msg *types.Message) error {
	// ISV 给指定租户发送消息
	resp, err := ch.client.Im.Message.Create(context.Background(), larkim.NewCreateMessageReqBuilder().
		ReceiveIdType(larkim.ReceiveIdTypeOpenId).
		Body(larkim.NewCreateMessageReqBodyBuilder().
			MsgType(larkim.MsgTypePost).
			ReceiveId("ou_c245b0a7dff2725cfa2fb104f8b48b9d").
			Content("text").
			Build()).
		Build(), larkcore.WithTenantKey(ch.tenantKey))

	if err != nil {
		return errors.Wrap(err, "reply_error")
	}

	log.
		WithField("message", msg).
		WithField("resp", resp).
		Info("reply ok")

	return nil
}
