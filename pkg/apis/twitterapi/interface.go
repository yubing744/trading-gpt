package twitterapi

import (
	"context"
	"fmt"
)

// Common error codes
const (
	ErrCodeUnauthorized   = 401
	ErrCodeForbidden      = 403
	ErrCodeRateLimit      = 429
	ErrCodeInternalError  = 500
	ErrCodeGatewayTimeout = 504
)

// HTTPError represents an HTTP error response
type HTTPError struct {
	StatusCode int
	Message    string
	Body       string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("http error: status=%d, message=%s", e.StatusCode, e.Message)
}

// TwitterAPIError represents an error response from the Twitter API
type TwitterAPIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *TwitterAPIError) Error() string {
	return fmt.Sprintf("twitterapi error: code=%d, message=%s", e.Code, e.Message)
}

// SearchRequest represents the payload for searching tweets
type SearchRequest struct {
	Query     string `json:"query"`      // Search query string (e.g., "AI", "from:elonmusk")
	QueryType string `json:"query_type"` // Search type ("Latest" or "Top", default: "Latest")
	Cursor    string `json:"cursor"`     // Pagination token for retrieving subsequent result pages (optional)
}

// Author represents the author of a tweet
type Author struct {
	ID              string `json:"id"`
	Username        string `json:"username"`
	Name            string `json:"name"`
	FollowerCount   int    `json:"follower_count"`
	FollowingCount  int    `json:"following_count"`
	Verified        bool   `json:"verified"`
	ProfileImageURL string `json:"profile_image_url"`
}

// Tweet represents a single tweet
type Tweet struct {
	ID           string   `json:"id"`
	URL          string   `json:"url"`
	Text         string   `json:"text"`
	Author       Author   `json:"author"`
	RetweetCount int      `json:"retweet_count"`
	LikeCount    int      `json:"like_count"`
	ReplyCount   int      `json:"reply_count"`
	ViewCount    int      `json:"view_count"`
	CreatedAt    string   `json:"created_at"`
	Language     string   `json:"lang"`
	Hashtags     []string `json:"hashtags"`
}

// SearchResponse represents the response from searching tweets
type SearchResponse struct {
	Tweets      []Tweet `json:"tweets"`
	HasNextPage bool    `json:"has_next_page"`
	NextCursor  string  `json:"next_cursor"`
}

// ITwitterClient is an interface for Twitter API client
type ITwitterClient interface {
	SearchTweets(ctx context.Context, req *SearchRequest) (*SearchResponse, error)
}
