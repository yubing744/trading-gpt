package feishu

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"github.com/yubing744/trading-bot/pkg/types"

	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

type FeishuChatChannel struct {
	id            string
	client        *lark.Client
	tenantKey     string
	receiveIdType string
	receiveId     string
	callbacks     []types.MessageCallback
}

func NewFeishuChatChannel(client *lark.Client, tenantKey string, receiveIdType string, receiveId string) *FeishuChatChannel {
	return &FeishuChatChannel{
		id:            fmt.Sprintf("%s:%s:%s", tenantKey, receiveIdType, receiveId),
		client:        client,
		tenantKey:     tenantKey,
		receiveIdType: receiveIdType,
		receiveId:     receiveId,
	}
}

func (ch *FeishuChatChannel) GetID() string {
	return ch.id
}

func (ch *FeishuChatChannel) toMessage(event *larkim.P2MessageReceiveV1) *types.Message {
	return &types.Message{
		ID:   event.EventReq.RequestId(),
		Text: *event.Event.Message.Content,
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
	content := map[string]string{
		"text": msg.Text,
	}
	contentBody, _ := json.Marshal(content)

	// ISV 给指定租户发送消息
	resp, err := ch.client.Im.Message.Create(context.Background(), larkim.NewCreateMessageReqBuilder().
		ReceiveIdType(ch.receiveIdType).
		Body(larkim.NewCreateMessageReqBodyBuilder().
			MsgType(larkim.MsgTypeText).
			ReceiveId(ch.receiveId).
			Content(string(contentBody)).
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
