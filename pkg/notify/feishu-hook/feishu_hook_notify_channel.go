package feishu_hook

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/yubing744/trading-bot/pkg/config"
	"github.com/yubing744/trading-bot/pkg/types"
)

var log = logrus.WithField("notify", "feishu_hook")

type FeishuHook struct {
	MsgType string         `json:"msg_type"`
	Content *types.Message `json:"content"`
}

type FeishuHookNotifyChannel struct {
	url    string
	client *http.Client
}

func NewFeishuHookNotifyChannel(cfg *config.NotifyFeishuHookConfig) *FeishuHookNotifyChannel {
	if cfg.URL == "" {
		log.Fatal("URL required!")
	}

	log.WithField("url", cfg.URL).
		Info("create feishu hook notify channel")

	return &FeishuHookNotifyChannel{
		url: cfg.URL,
		client: &http.Client{
			Timeout: time.Second * 20,
		},
	}
}

func (ch *FeishuHookNotifyChannel) GetID() string {
	return "feishu_hook"
}

func (ch *FeishuHookNotifyChannel) Reply(ctx context.Context, msg *types.Message) error {
	hook := &FeishuHook{
		MsgType: "text",
		Content: msg,
	}

	body, _ := json.Marshal(hook)
	req, err := http.NewRequest("POST", ch.url, bytes.NewReader(body))
	if err != nil {
		return err
	}

	log.WithField("url", ch.url).
		WithField("body", string(body)).
		Debug("reply")

	resp, err := ch.client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return errors.Errorf("response error, status code: %d, detail: %s", resp.StatusCode, body)
	}

	return nil
}
