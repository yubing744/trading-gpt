package twitterapi

import (
	"context"
	"os"
	"testing"
	"time"
)

// TestSearchTweets_Integration tests the real API call
// Run with: go test -v -run TestSearchTweets_Integration ./pkg/apis/twitterapi/
func TestSearchTweets_Integration(t *testing.T) {
	apiKey := os.Getenv("TWITTER_API_KEY")
	if apiKey == "" {
		t.Skip("TWITTER_API_KEY not set, skipping integration test")
	}

	client := NewClient("https://api.twitterapi.io", apiKey, WithTimeout(30*time.Second))

	ctx := context.Background()
	req := &SearchRequest{
		Query:     "SUI",
		QueryType: "Latest",
	}

	t.Logf("Searching for tweets with query: %s", req.Query)

	resp, err := client.SearchTweets(ctx, req)
	if err != nil {
		t.Fatalf("SearchTweets failed: %v", err)
	}

	t.Logf("Found %d tweets", len(resp.Tweets))
	t.Logf("Has next page: %v", resp.HasNextPage)
	if resp.NextCursor != "" {
		t.Logf("Next cursor: %s", resp.NextCursor)
	}

	for i, tweet := range resp.Tweets {
		if i >= 3 { // Only log first 3 tweets
			break
		}
		t.Logf("\nTweet %d:", i+1)
		t.Logf("  Author: @%s (%s)", tweet.Author.Username, tweet.Author.Name)
		t.Logf("  Text: %s", tweet.Text)
		t.Logf("  Likes: %d, Retweets: %d, Replies: %d", tweet.LikeCount, tweet.RetweetCount, tweet.ReplyCount)
		t.Logf("  Created: %s", tweet.CreatedAt)
		t.Logf("  URL: %s", tweet.URL)
	}

	if len(resp.Tweets) == 0 {
		t.Log("No tweets found, but request was successful")
	}
}
