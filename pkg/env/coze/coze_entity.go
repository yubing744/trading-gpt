package coze

import (
	"context"
	"strings"
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
	client := coze.NewClient(config.BaseURL, config.APIKey, coze.WithTimeout(config.Timeout.Duration()))

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

		interval := item.Interval.Duration()
		nextTick := time.Now().Truncate(interval).Add(interval)
		initialDelay := time.Until(nextTick) - item.Before.Duration()

		time.AfterFunc(initialDelay, func() {
			e.communicateWithCozeBot(ctx, ch, item)
			ticker := time.NewTicker(interval)
			e.timers[item.Name] = ticker

			go func(item *config.IndicatorItem, ticker *time.Ticker) {
				for {
					select {
					case <-ctx.Done():
						ticker.Stop()
						return
					case <-ticker.C:
						e.communicateWithCozeBot(ctx, ch, item)
					}
				}
			}(item, ticker)
		})
	}

	for _, item := range e.config.WorkflowIndicatorItems {
		log.WithField("item", item).Info("coze_run_workflow_item")

		interval := item.Interval.Duration()
		nextTick := time.Now().Truncate(interval).Add(interval)
		initialDelay := time.Until(nextTick) - item.Before.Duration()

		time.AfterFunc(initialDelay, func() {
			e.communicateWithCozeWorkflow(ctx, ch, item)
			ticker := time.NewTicker(interval)
			e.timers[item.Name] = ticker

			go func(item *config.WorkflowIndicatorItem, ticker *time.Ticker) {
				for {
					select {
					case <-ctx.Done():
						ticker.Stop()
						return
					case <-ticker.C:
						e.communicateWithCozeWorkflow(ctx, ch, item)
					}
				}
			}(item, ticker)
		})
	}

	<-ctx.Done() // Wait for cancellation
}

// communicateWithCozeBot handles the communication with the Coze platform for a given scheduled task and sends events to the channel.
func (e *CozeEntity) communicateWithCozeBot(ctx context.Context, ch chan types.IEvent, item *config.IndicatorItem) {
	req := &coze.ChatRequest{
		ConversationID: uuid.NewString(),
		BotID:          item.BotID,
		Query:          item.Message,
		User:           "29032201862555",
		Stream:         false,
	}

	log.WithField("item", item).WithField("req", req).Info("communicateWithCozeBot_start")

	response, err := e.cozeClient.Chat(ctx, req)
	if err != nil {
		log.WithField("item", item).WithField("req", req).WithError(err).Error("communicateWithCozeBot_error")
		return
	}

	log.WithField("item", item).WithField("req", req).WithField("response", response).Info("communicateWithCozeBot_end")

	if response.Code != 0 {
		log.WithField("errorMsg", response.Msg).Error("response_error")
		return
	}

	sb := strings.Builder{}

	for _, msg := range response.Messages {
		if msg.Role == "assistant" && msg.Type == "answer" {
			sb.WriteString(msg.Content)
		}
	}

	event := NewCozeEvent(item.Name, item.Description, sb.String())
	ch <- event
}

// communicateWithCoze handles the communication with the Coze platform for a given scheduled task and sends events to the channel.
func (e *CozeEntity) communicateWithCozeWorkflow(ctx context.Context, ch chan types.IEvent, item *config.WorkflowIndicatorItem) {
	req := &coze.WorkflowRequest{
		WorkflowID: item.WorkflowID,
		Parameters: item.Params,
	}

	log.WithField("item", item).WithField("req", req).Info("communicateWithCozeWorkflow_start")

	resp, err := e.cozeClient.RunWorkflow(ctx, req)
	if err != nil {
		log.WithField("item", item).WithField("req", req).WithError(err).Error("communicateWithCozeWorkflow_error")
		return
	}

	log.WithField("item", item).WithField("req", req).WithField("resp", resp).Info("communicateWithCozeWorkflow_end")

	if resp.Code != 0 {
		log.WithField("errorCode", resp.Code).WithField("errorMsg", resp.Msg).Error("response_error")
		return
	}

	event := NewCozeEvent(item.Name, item.Description, resp.Data)
	ch <- event
}
