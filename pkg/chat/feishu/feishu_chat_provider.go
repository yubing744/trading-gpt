package feishu

import (
	"context"
	"fmt"
	"io"
	"net/http"

	lark "github.com/larksuite/oapi-sdk-go/v3"
	"github.com/sirupsen/logrus"
	"github.com/yubing744/trading-bot/pkg/config"

	"github.com/kataras/go-events"

	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	"github.com/larksuite/oapi-sdk-go/v3/core/httpserverext"
	larkevent "github.com/larksuite/oapi-sdk-go/v3/event"
	"github.com/larksuite/oapi-sdk-go/v3/event/dispatcher"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

var log = logrus.WithField("chat", "feishu")

type FeishuChatProvider struct {
	events.EventEmmiter
	client     *lark.Client
	dispatcher *dispatcher.EventDispatcher
	serverPort int
}

func NewFeishuChatProvider(cfg *config.ChatFeishuConfig) *FeishuChatProvider {
	// 创建 API Client
	var cli = lark.NewClient(cfg.AppId, cfg.AppSecret, lark.WithLogReqAtDebug(true), lark.WithLogLevel(larkcore.LogLevelDebug))

	eventEncryptKey := cfg.EventEncryptKey
	verificationToken := cfg.VerificationToken
	var dispatcher = dispatcher.NewEventDispatcher(verificationToken, eventEncryptKey)

	return &FeishuChatProvider{
		EventEmmiter: events.New(),
		client:       cli,
		dispatcher:   dispatcher,
		serverPort:   cfg.ServerPort,
	}
}

func (feishu *FeishuChatProvider) Start() error {
	// 注册消息处理器
	handler := feishu.dispatcher.OnP2MessageReceiveV1(func(ctx context.Context, event *larkim.P2MessageReceiveV1) error {
		// 处理消息 event，这里简单打印消息的内容
		fmt.Println(larkcore.Prettify(event))
		fmt.Println(event.RequestId())

		// 获取租户 key 并发送消息
		tenantKey := event.TenantKey()

		// ISV 给指定租户发送消息
		resp, err := feishu.client.Im.Message.Create(context.Background(), larkim.NewCreateMessageReqBuilder().
			ReceiveIdType(larkim.ReceiveIdTypeOpenId).
			Body(larkim.NewCreateMessageReqBodyBuilder().
				MsgType(larkim.MsgTypePost).
				ReceiveId("ou_c245b0a7dff2725cfa2fb104f8b48b9d").
				Content("text").
				Build()).
			Build(), larkcore.WithTenantKey(tenantKey))

		// 发送结果处理，resp,err
		fmt.Println(resp, err)

		return nil
	})

	mux := http.NewServeMux()

	// 注册 http 路由
	mux.HandleFunc("/ping", func(w http.ResponseWriter, req *http.Request) {
		io.WriteString(w, "ok")
	})

	mux.HandleFunc("/webhook/event", httpserverext.NewEventHandlerFunc(handler, larkevent.WithLogLevel(larkcore.LogLevelDebug)))

	// 启动 http 服务
	port := fmt.Sprintf(":%d", feishu.serverPort)
	log.Infof("start chat feishu at %s ok", port)
	err := http.ListenAndServe(port, mux)
	return err
}
