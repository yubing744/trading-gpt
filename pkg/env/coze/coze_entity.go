package coze

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/yubing744/trading-gpt/pkg/apis/coze"
	"github.com/yubing744/trading-gpt/pkg/config"
	"github.com/yubing744/trading-gpt/pkg/types"
)

var log = logrus.WithField("entity", "coze")

type CozeEntity struct {
	id         string
	cozeClient coze.ICozeClient
	config     *config.CozeEntityConfig
	timers     map[string]*time.Ticker
}

// NewCozeEntity creates a new instance of CozeEntity with the given ID, Coze client, and configuration.
func NewCozeEntity(config *config.CozeEntityConfig) *CozeEntity {
	client := coze.NewClient(config.BaseURL, config.APIKey, config.Timeout.Duration())

	return &CozeEntity{
		id:         "coze",
		cozeClient: client,
		config:     config,
		timers:     make(map[string]*time.Ticker),
	}
}

// GetID returns the entity's id.
func (e *CozeEntity) GetID() string {
	return e.id
}

// Actions returns a list of action descriptors.
func (e *CozeEntity) Actions() []*types.ActionDesc {
	// TODO: Return the actual actions this entity can perform.
	return []*types.ActionDesc{}
}

// HandleCommand handles a command directed at the entity.
func (e *CozeEntity) HandleCommand(ctx context.Context, cmd string, args map[string]string) error {
	// TODO: Implement the logic to handle the command using CozeClient.
	return nil
}

// Run starts the entity's main loop and sets up scheduled tasks based on the entity's configuration.
func (e *CozeEntity) Run(ctx context.Context, ch chan types.IEvent) {
	log.Info("coze_run")

	for _, item := range e.config.IndicatorItems {
		log.WithField("item", item).Info("coze_run_item")
		e.communicateWithCoze(ctx, ch, item)

		ticker := time.NewTicker(item.Interval.Duration())
		e.timers[item.Name] = ticker
		go func(item *config.IndicatorItem, ticker *time.Ticker) {
			for {
				select {
				case <-ctx.Done():
					ticker.Stop()
					return
				case <-ticker.C:
					e.communicateWithCoze(ctx, ch, item)
				}
			}
		}(item, ticker)
	}
	<-ctx.Done() // Wait for cancellation
}

// communicateWithCoze handles the communication with the Coze platform for a given scheduled task and sends events to the channel.
func (e *CozeEntity) communicateWithCoze(ctx context.Context, ch chan types.IEvent, item *config.IndicatorItem) {

	req := &coze.ChatRequest{
		ConversationID: uuid.NewString(),
		BotID:          item.BotID,
		Query:          item.Message,
		User:           "29032201862555",
		Stream:         false,
	}

	log.WithField("item", item).WithField("req", req).Info("communicateWithCoze_start")

	response, err := e.cozeClient.Chat(ctx, req)
	if err != nil {
		log.WithField("item", item).WithField("req", req).WithError(err).Error("communicateWithCoze_error")
		return
	}

	log.WithField("item", item).WithField("req", req).WithField("response", response).Info("communicateWithCoze_end")

	if response.Code != 0 {
		log.WithField("errorMsg", response.Msg).Error("response_error")
		return
	}

	event := NewCozeEvent(item, response)
	ch <- event
}
