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

// Actions returns a list of action descriptors.
func (e *TwitterAPIEntity) Actions() []*types.ActionDesc {
	return []*types.ActionDesc{}
}

// HandleCommand handles a command directed at the entity.
func (e *TwitterAPIEntity) HandleCommand(ctx context.Context, cmd string, args map[string]string) error {
	return nil
}

// Run starts the entity's main loop and sets up scheduled tasks based on the entity's configuration.
func (e *TwitterAPIEntity) Run(ctx context.Context, ch chan types.IEvent) {
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

	for i := 0; i < count; i++ {
		tweet := tweets[i]
		sb.WriteString(fmt.Sprintf("Tweet %d:\n", i+1))
		sb.WriteString(fmt.Sprintf("Author: @%s (%s)\n", tweet.Author.Username, tweet.Author.Name))
		sb.WriteString(fmt.Sprintf("Text: %s\n", tweet.Text))
		sb.WriteString(fmt.Sprintf("Engagement: %d likes, %d retweets, %d replies\n", tweet.LikeCount, tweet.RetweetCount, tweet.ReplyCount))
		sb.WriteString(fmt.Sprintf("Created: %s\n", tweet.CreatedAt))
		sb.WriteString(fmt.Sprintf("URL: %s\n", tweet.URL))

		if len(tweet.Hashtags) > 0 {
			sb.WriteString(fmt.Sprintf("Hashtags: %s\n", strings.Join(tweet.Hashtags, ", ")))
		}

		sb.WriteString("\n---\n\n")
	}

	if len(sb.String()) == 0 {
		return "No tweets found."
	}

	return sb.String()
}
