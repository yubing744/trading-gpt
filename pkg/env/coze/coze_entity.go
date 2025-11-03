package coze

import (
	"context"
	"fmt"
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

// Actions returns a list of action descriptors for available workflows.
func (e *CozeEntity) Actions() []*types.ActionDesc {
	actions := make([]*types.ActionDesc, 0)

	// Add workflow actions from configuration
	for _, workflow := range e.config.WorkflowIndicatorItems {
		action := &types.ActionDesc{
			Name:        workflow.Name,
			Description: workflow.Description,
			Args: []types.ArgmentDesc{
				{
					Name:        "workflow_id",
					Description: "Coze workflow ID (optional, uses configured ID if not specified)",
				},
			},
		}

		// Add dynamic parameters if workflow has them
		if len(workflow.Params) > 0 {
			for key := range workflow.Params {
				action.Args = append(action.Args, types.ArgmentDesc{
					Name:        key,
					Description: fmt.Sprintf("Parameter %s for workflow", key),
				})
			}
		}

		actions = append(actions, action)
	}

	// Add bot actions from configuration
	for _, bot := range e.config.IndicatorItems {
		action := &types.ActionDesc{
			Name:        bot.Name,
			Description: bot.Description,
			Args: []types.ArgmentDesc{
				{
					Name:        "message",
					Description: "Message to send to Coze bot (optional, uses configured message if not specified)",
				},
			},
		}
		actions = append(actions, action)
	}

	return actions
}

// HandleCommand handles a command directed at the entity.
func (e *CozeEntity) HandleCommand(ctx context.Context, cmd string, args map[string]string) error {
	// Check if command matches a workflow
	for _, workflow := range e.config.WorkflowIndicatorItems {
		if workflow.Name == cmd {
			return e.executeWorkflowCommand(ctx, workflow, args)
		}
	}

	// Check if command matches a bot
	for _, bot := range e.config.IndicatorItems {
		if bot.Name == cmd {
			return e.executeBotCommand(ctx, bot, args)
		}
	}

	return fmt.Errorf("unknown command: %s", cmd)
}

// executeWorkflowCommand executes a Coze workflow command
func (e *CozeEntity) executeWorkflowCommand(ctx context.Context, workflow *config.WorkflowIndicatorItem, args map[string]string) error {
	// Use workflow ID from args if provided, otherwise use configured ID
	workflowID := workflow.WorkflowID
	if providedID, ok := args["workflow_id"]; ok && providedID != "" {
		workflowID = providedID
	}

	// Merge args with configured params (args take precedence)
	params := make(map[string]string)
	for k, v := range workflow.Params {
		params[k] = v
	}
	for k, v := range args {
		if k != "workflow_id" { // Skip the special workflow_id parameter
			params[k] = v
		}
	}

	// Execute workflow
	req := &coze.WorkflowRequest{
		WorkflowID: workflowID,
		Parameters: params,
	}

	log.WithField("workflowID", workflowID).WithField("params", params).Info("Executing workflow command")

	resp, err := e.cozeClient.RunWorkflow(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to run workflow: %w", err)
	}

	if resp.Code != 0 {
		return fmt.Errorf("workflow execution failed: %s (code: %d)", resp.Msg, resp.Code)
	}

	log.WithField("workflowID", workflowID).WithField("response", resp.Data).Info("Workflow command executed successfully")
	return nil
}

// executeBotCommand executes a Coze bot command
func (e *CozeEntity) executeBotCommand(ctx context.Context, bot *config.IndicatorItem, args map[string]string) error {
	// Use message from args if provided, otherwise use configured message
	message := bot.Message
	if providedMsg, ok := args["message"]; ok && providedMsg != "" {
		message = providedMsg
	}

	// Execute bot chat
	req := &coze.ChatRequest{
		ConversationID: uuid.NewString(),
		BotID:          bot.BotID,
		Query:          message,
		User:           "29032201862555",
		Stream:         false,
	}

	log.WithField("botID", bot.BotID).WithField("message", message).Info("Executing bot command")

	response, err := e.cozeClient.Chat(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to execute bot chat: %w", err)
	}

	if response.Code != 0 {
		return fmt.Errorf("bot execution failed: %s (code: %d)", response.Msg, response.Code)
	}

	log.WithField("botID", bot.BotID).Info("Bot command executed successfully")
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
