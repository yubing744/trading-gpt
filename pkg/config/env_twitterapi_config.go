package config

import (
	"github.com/c9s/bbgo/pkg/types"
)

// TwitterAPISearchItem represents a single scheduled search task with its configuration.
type TwitterAPISearchItem struct {
	Name        string         `json:"name"`        // A unique name for the search task
	Description string         `json:"description"` // A description of what the search does
	Query       string         `json:"query"`       // Search query string (e.g., "AI", "from:elonmusk")
	QueryType   string         `json:"query_type"`  // Search type ("Latest" or "Top", default: "Latest")
	Interval    types.Interval `json:"interval"`    // How often to run the search
	Before      types.Interval `json:"before"`      // Offset before interval
	MaxResults  int            `json:"max_results"` // Max tweets per search (default: 20)
}

// TwitterAPIEntityConfig holds the configuration for a TwitterAPIEntity.
type TwitterAPIEntityConfig struct {
	Enabled     bool                    `json:"enabled"`
	BaseURL     string                  `json:"base_url"` // Default: https://api.twitterapi.io
	APIKey      string                  `json:"api_key"`
	Timeout     types.Interval          `json:"timeout"`
	SearchItems []*TwitterAPISearchItem `json:"search_items"` // A list of scheduled search tasks
}
