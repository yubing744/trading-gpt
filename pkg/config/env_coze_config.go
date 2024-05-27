package config

import (
	"github.com/c9s/bbgo/pkg/types"
)

// IndicatorItem represents a single scheduled task with its configuration.
type IndicatorItem struct {
	Name        string         `json:"name"`        // A unique name for the scheduled task
	Description string         `json:"description"` // A description of what the task does
	Interval    types.Interval `json:"interval"`    // How often to run the task
	Before      types.Interval `json:"before"`      // How often to run the task
	BotID       string         `json:"bot_id"`      // The ID of the bot to interact with
	Message     string         `json:"message"`     // The message content to send to the bot
}

// CozeEntityConfig holds the configuration for a CozeEntity.
type CozeEntityConfig struct {
	Enabled        bool             `json:"enabled"`
	BaseURL        string           `json:"base_url"`
	APIKey         string           `json:"api_key"`
	Timeout        types.Interval   `json:"timeout"`
	IndicatorItems []*IndicatorItem `json:"indicators"` // A list of scheduled tasks
}
