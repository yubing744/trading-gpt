package feishu_hook

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/yubing744/trading-bot/pkg/config"
	"github.com/yubing744/trading-bot/pkg/types"
)

var log = logrus.WithField("notify", "feishu_hook")

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
	body := fmt.Sprintf("{\"msg_type\":\"text\",\"content\":{\"text\":\"%s\"}}", msg.Text)

	req, err := http.NewRequest("POST", ch.url, strings.NewReader(body))
	if err != nil {
		return err
	}

	log.WithField("url", ch.url).
		WithField("body", body).
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
