package twitterapi

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/yubing744/trading-gpt/pkg/apis/twitterapi"
	"github.com/yubing744/trading-gpt/pkg/config"
	"github.com/yubing744/trading-gpt/pkg/types"
)

var log = logrus.WithField("entity", "twitterapi")

type TwitterAPIEntity struct {
	id            string
	twitterClient twitterapi.ITwitterClient
	config        *config.TwitterAPIEntityConfig
	timers        map[string]*time.Ticker
	eventChannel  chan types.IEvent // Store event channel for command execution
}

// NewTwitterAPIEntity creates a new instance of TwitterAPIEntity with the given ID, Twitter client, and configuration.
func NewTwitterAPIEntity(config *config.TwitterAPIEntityConfig) *TwitterAPIEntity {
	client := twitterapi.NewClient(config.BaseURL, config.APIKey, twitterapi.WithTimeout(config.Timeout.Duration()))

	return &TwitterAPIEntity{
		id:            "twitterapi",
		twitterClient: client,
		config:        config,
		timers:        make(map[string]*time.Ticker),
	}
}

// GetID returns the entity's id.
func (e *TwitterAPIEntity) GetID() string {
	return e.id
}

// Actions returns a list of action descriptors for available Twitter searches.
func (e *TwitterAPIEntity) Actions() []*types.ActionDesc {
	actions := make([]*types.ActionDesc, 0)

	// Add configured search items as available actions
	for _, item := range e.config.SearchItems {
		action := &types.ActionDesc{
			Name:        item.Name,
			Description: item.Description,
			Args: []types.ArgmentDesc{
				{
					Name:        "query",
					Description: "Search query (optional, uses configured query if not specified)",
				},
				{
					Name:        "query_type",
					Description: "Query type: Top|Latest (optional, uses configured type if not specified)",
				},
				{
					Name:        "max_results",
					Description: "Maximum number of results to return (optional, uses configured max if not specified)",
				},
			},
		}
		actions = append(actions, action)
	}

	// Add a generic search action
	actions = append(actions, &types.ActionDesc{
		Name:        "search_tweets",
		Description: "Search Twitter for tweets matching a query",
		Args: []types.ArgmentDesc{
			{
				Name:        "query",
				Description: "Search query (required)",
			},
			{
				Name:        "query_type",
				Description: "Query type: Top|Latest (default: Top)",
			},
			{
				Name:        "max_results",
				Description: "Maximum number of results to return (default: 10, max: 100)",
			},
		},
	})

	return actions
}

// HandleCommand handles a command directed at the entity.
func (e *TwitterAPIEntity) HandleCommand(ctx context.Context, cmd string, args map[string]string) error {
	if e.eventChannel == nil {
		return fmt.Errorf("event channel not initialized, command can only be executed during Run()")
	}

	// Check if command matches a configured search item
	for _, item := range e.config.SearchItems {
		if item.Name == cmd {
			return e.executeConfiguredSearch(ctx, item, args)
		}
	}

	// Check if command is the generic search_tweets
	if cmd == "search_tweets" {
		return e.executeGenericSearch(ctx, args)
	}

	return fmt.Errorf("unknown command: %s", cmd)
}

// executeConfiguredSearch executes a configured search item with optional parameter overrides
func (e *TwitterAPIEntity) executeConfiguredSearch(ctx context.Context, item *config.TwitterAPISearchItem, args map[string]string) error {
	// Use configured values as defaults, override with args if provided
	query := item.Query
	if providedQuery, ok := args["query"]; ok && providedQuery != "" {
		query = providedQuery
	}

	queryType := item.QueryType
	if providedType, ok := args["query_type"]; ok && providedType != "" {
		queryType = providedType
	}

	maxResults := item.MaxResults
	if providedMax, ok := args["max_results"]; ok && providedMax != "" {
		parsedMax, err := fmt.Sscanf(providedMax, "%d", &maxResults)
		if err != nil || parsedMax != 1 {
			return fmt.Errorf("invalid max_results parameter: %s", providedMax)
		}
	}

	log.WithField("query", query).WithField("queryType", queryType).Info("Executing configured Twitter search command")

	// Execute search
	req := &twitterapi.SearchRequest{
		Query:     query,
		QueryType: queryType,
	}

	response, err := e.twitterClient.SearchTweets(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to search tweets: %w", err)
	}

	// Format and send results
	content := e.formatTweets(response.Tweets, maxResults)
	event := NewTwitterAPIEvent(item.Name, item.Description, content)
	e.eventChannel <- event

	log.WithField("tweetCount", len(response.Tweets)).Info("Twitter search command executed successfully")
	return nil
}

// executeGenericSearch executes a generic Twitter search with parameters from args
func (e *TwitterAPIEntity) executeGenericSearch(ctx context.Context, args map[string]string) error {
	// Query is required
	query, ok := args["query"]
	if !ok || query == "" {
		return fmt.Errorf("query parameter is required for search_tweets command")
	}

	// Query type defaults to "Top"
	queryType := "Top"
	if providedType, ok := args["query_type"]; ok && providedType != "" {
		queryType = providedType
	}

	// Max results defaults to 10
	maxResults := 10
	if providedMax, ok := args["max_results"]; ok && providedMax != "" {
		parsedMax, err := fmt.Sscanf(providedMax, "%d", &maxResults)
		if err != nil || parsedMax != 1 {
			return fmt.Errorf("invalid max_results parameter: %s", providedMax)
		}
		if maxResults > 100 {
			maxResults = 100
		}
		if maxResults < 1 {
			maxResults = 1
		}
	}

	log.WithField("query", query).WithField("queryType", queryType).Info("Executing generic Twitter search command")

	// Execute search
	req := &twitterapi.SearchRequest{
		Query:     query,
		QueryType: queryType,
	}

	response, err := e.twitterClient.SearchTweets(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to search tweets: %w", err)
	}

	// Format and send results
	content := e.formatTweets(response.Tweets, maxResults)
	description := fmt.Sprintf("Twitter search results for: %s", query)
	event := NewTwitterAPIEvent("search_tweets", description, content)
	e.eventChannel <- event

	log.WithField("tweetCount", len(response.Tweets)).Info("Generic Twitter search command executed successfully")
	return nil
}

// Run starts the entity's main loop and sets up scheduled tasks based on the entity's configuration.
func (e *TwitterAPIEntity) Run(ctx context.Context, ch chan types.IEvent) {
	// Store event channel for command execution
	e.eventChannel = ch

	log.Info("twitterapi_run")

	for _, item := range e.config.SearchItems {
		log.WithField("item", item).Info("twitterapi_run_item")

		interval := item.Interval.Duration()
		nextTick := time.Now().Truncate(interval).Add(interval)
		initialDelay := time.Until(nextTick) - item.Before.Duration()

		time.AfterFunc(initialDelay, func() {
			e.searchTweets(ctx, ch, item)
			ticker := time.NewTicker(interval)
			e.timers[item.Name] = ticker

			go func(item *config.TwitterAPISearchItem, ticker *time.Ticker) {
				for {
					select {
					case <-ctx.Done():
						ticker.Stop()
						return
					case <-ticker.C:
						e.searchTweets(ctx, ch, item)
					}
				}
			}(item, ticker)
		})
	}

	<-ctx.Done() // Wait for cancellation
}

// searchTweets handles the Twitter search for a given scheduled task and sends events to the channel.
func (e *TwitterAPIEntity) searchTweets(ctx context.Context, ch chan types.IEvent, item *config.TwitterAPISearchItem) {
	req := &twitterapi.SearchRequest{
		Query:     item.Query,
		QueryType: item.QueryType,
	}

	log.WithField("item", item).WithField("req", req).Info("searchTweets_start")

	response, err := e.twitterClient.SearchTweets(ctx, req)
	if err != nil {
		log.WithField("item", item).WithField("req", req).WithError(err).Error("searchTweets_error")
		return
	}

	log.WithField("item", item).WithField("req", req).WithField("response", response).Info("searchTweets_end")

	// Format tweets into a readable content
	content := e.formatTweets(response.Tweets, item.MaxResults)

	event := NewTwitterAPIEvent(item.Name, item.Description, content)
	ch <- event
}

// formatTweets formats the tweets into a human-readable string
func (e *TwitterAPIEntity) formatTweets(tweets []twitterapi.Tweet, maxResults int) string {
	sb := strings.Builder{}

	// Limit the number of tweets if maxResults is specified
	count := len(tweets)
	if maxResults > 0 && maxResults < count {
		count = maxResults
	}

	now := time.Now()

	for i := 0; i < count; i++ {
		tweet := tweets[i]
		sb.WriteString(fmt.Sprintf("Tweet %d:\n", i+1))
		sb.WriteString(fmt.Sprintf("Author: @%s (%s)\n", tweet.Author.UserName, tweet.Author.Name))
		sb.WriteString(fmt.Sprintf("Text: %s\n", tweet.Text))
		sb.WriteString(fmt.Sprintf("Engagement: %d likes, %d retweets, %d replies\n", tweet.LikeCount, tweet.RetweetCount, tweet.ReplyCount))

		// Parse and format creation time with relative time
		relativeTime := e.formatRelativeTime(tweet.CreatedAt, now)
		sb.WriteString(fmt.Sprintf("Created: %s (%s)\n", tweet.CreatedAt, relativeTime))

		sb.WriteString(fmt.Sprintf("URL: %s\n", tweet.URL))

		if len(tweet.Entities.Hashtags) > 0 {
			hashtags := make([]string, len(tweet.Entities.Hashtags))
			for i, h := range tweet.Entities.Hashtags {
				hashtags[i] = h.Text
			}
			sb.WriteString(fmt.Sprintf("Hashtags: %s\n", strings.Join(hashtags, ", ")))
		}

		sb.WriteString("\n---\n\n")
	}

	if len(sb.String()) == 0 {
		return "No tweets found."
	}

	return sb.String()
}

// formatRelativeTime converts a timestamp to relative time (e.g., "5 minutes ago", "2 hours ago")
func (e *TwitterAPIEntity) formatRelativeTime(createdAt string, now time.Time) string {
	// Twitter time format: "Mon Jan 02 15:04:05 -0700 2006"
	const twitterTimeFormat = "Mon Jan 02 15:04:05 -0700 2006"

	t, err := time.Parse(twitterTimeFormat, createdAt)
	if err != nil {
		// Try RFC3339 format as fallback
		t, err = time.Parse(time.RFC3339, createdAt)
		if err != nil {
			// Try alternative format
			t, err = time.Parse("2006-01-02T15:04:05Z", createdAt)
			if err != nil {
				return "unknown"
			}
		}
	}

	duration := now.Sub(t)

	// Format based on duration
	switch {
	case duration < time.Minute:
		seconds := int(duration.Seconds())
		if seconds <= 1 {
			return "just now"
		}
		return fmt.Sprintf("%d seconds ago", seconds)
	case duration < time.Hour:
		minutes := int(duration.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	case duration < 24*time.Hour:
		hours := int(duration.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	case duration < 7*24*time.Hour:
		days := int(duration.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	case duration < 30*24*time.Hour:
		weeks := int(duration.Hours() / 24 / 7)
		if weeks == 1 {
			return "1 week ago"
		}
		return fmt.Sprintf("%d weeks ago", weeks)
	case duration < 365*24*time.Hour:
		months := int(duration.Hours() / 24 / 30)
		if months == 1 {
			return "1 month ago"
		}
		return fmt.Sprintf("%d months ago", months)
	default:
		years := int(duration.Hours() / 24 / 365)
		if years == 1 {
			return "1 year ago"
		}
		return fmt.Sprintf("%d years ago", years)
	}
}
