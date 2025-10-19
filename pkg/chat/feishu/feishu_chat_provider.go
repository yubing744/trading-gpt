package feishu

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/kataras/go-events"
	"github.com/sirupsen/logrus"
	"github.com/yubing744/trading-gpt/pkg/chat"
	"github.com/yubing744/trading-gpt/pkg/config"

	lark "github.com/larksuite/oapi-sdk-go/v3"
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
	channels   map[string]*FeishuChatChannel
}

func NewFeishuChatProvider(cfg *config.ChatFeishuConfig) *FeishuChatProvider {
	// Create API Client
	opts := []lark.ClientOptionFunc{}
	debugMock := os.Getenv("DEBUG")
	if debugMock == "true" {
		opts = append(opts, lark.WithLogReqAtDebug(true))
		opts = append(opts, lark.WithLogLevel(larkcore.LogLevelDebug))
	}
	var cli = lark.NewClient(cfg.AppId, cfg.AppSecret, opts...)

	eventEncryptKey := cfg.EventEncryptKey
	verificationToken := cfg.VerificationToken
	var dispatcher = dispatcher.NewEventDispatcher(verificationToken, eventEncryptKey)

	return &FeishuChatProvider{
		EventEmmiter: events.New(),
		client:       cli,
		dispatcher:   dispatcher,
		serverPort:   cfg.ServerPort,
		channels:     make(map[string]*FeishuChatChannel, 0),
	}
}

func (feishu *FeishuChatProvider) Listen(cb chat.ListenCallback) error {
	handler := feishu.dispatcher.OnP2MessageReceiveV1(func(ctx context.Context, event *larkim.P2MessageReceiveV1) error {
		log.Info(larkcore.Prettify(event))

		tenantKey := event.TenantKey()
		receiveIdType := larkim.ReceiveIdTypeOpenId
		receiveId := *event.Event.Sender.SenderId.OpenId

		if event.Event.Message != nil && *event.Event.Message.ChatType == "group" {
			receiveIdType = larkim.ReceiveIdTypeChatId
			receiveId = *event.Event.Message.ChatId
		}

		// Create channel
		channelKey := fmt.Sprintf("%s:%s:%s", tenantKey, receiveIdType, receiveId)
		channel, ok := feishu.channels[channelKey]
		if !ok {
			channel = NewFeishuChatChannel(feishu.client, tenantKey, receiveIdType, receiveId)
			feishu.channels[channelKey] = channel
			cb(channel)
		}

		go func() {
			channel.handleEvent(event)
		}()

		return nil
	})

	mux := http.NewServeMux()

	// Register HTTP routes
	mux.HandleFunc("/ping", func(w http.ResponseWriter, req *http.Request) {
		io.WriteString(w, "ok")
	})

	mux.HandleFunc("/webhook/event", httpserverext.NewEventHandlerFunc(handler, larkevent.WithLogLevel(larkcore.LogLevelDebug)))

	// Start HTTP service
	port := fmt.Sprintf(":%d", feishu.serverPort)
	log.Infof("start chat feishu at %s ok", port)
	err := http.ListenAndServe(port, mux)

	return err
}
