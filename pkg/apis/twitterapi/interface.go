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
	Type                       string   `json:"type"`
	UserName                   string   `json:"userName"`
	URL                        string   `json:"url"`
	ID                         string   `json:"id"`
	Name                       string   `json:"name"`
	IsBlueVerified             bool     `json:"isBlueVerified"`
	VerifiedType               string   `json:"verifiedType"`
	ProfilePicture             string   `json:"profilePicture"`
	CoverPicture               string   `json:"coverPicture"`
	Description                string   `json:"description"`
	Location                   string   `json:"location"`
	Followers                  int      `json:"followers"`
	Following                  int      `json:"following"`
	CanDm                      bool     `json:"canDm"`
	CreatedAt                  string   `json:"createdAt"`
	FavouritesCount            int      `json:"favouritesCount"`
	HasCustomTimelines         bool     `json:"hasCustomTimelines"`
	IsTranslator               bool     `json:"isTranslator"`
	MediaCount                 int      `json:"mediaCount"`
	StatusesCount              int      `json:"statusesCount"`
	WithheldInCountries        []string `json:"withheldInCountries"`
	PossiblySensitive          bool     `json:"possiblySensitive"`
	PinnedTweetIds             []string `json:"pinnedTweetIds"`
	IsAutomated                bool     `json:"isAutomated"`
	AutomatedBy                string   `json:"automatedBy"`
	Unavailable                bool     `json:"unavailable"`
	Message                    string   `json:"message"`
	UnavailableReason          string   `json:"unavailableReason"`
}

// Hashtag represents a hashtag in a tweet
type Hashtag struct {
	Indices []int  `json:"indices"`
	Text    string `json:"text"`
}

// URL represents a URL in a tweet
type URL struct {
	DisplayURL  string `json:"display_url"`
	ExpandedURL string `json:"expanded_url"`
	Indices     []int  `json:"indices"`
	URL         string `json:"url"`
}

// UserMention represents a user mention in a tweet
type UserMention struct {
	IDStr      string `json:"id_str"`
	Name       string `json:"name"`
	ScreenName string `json:"screen_name"`
}

// Entities represents the entities in a tweet
type Entities struct {
	Hashtags     []Hashtag     `json:"hashtags"`
	URLs         []URL         `json:"urls"`
	UserMentions []UserMention `json:"user_mentions"`
}

// Tweet represents a single tweet
type Tweet struct {
	Type              string   `json:"type"`
	ID                string   `json:"id"`
	URL               string   `json:"url"`
	Text              string   `json:"text"`
	Source            string   `json:"source"`
	RetweetCount      int      `json:"retweetCount"`
	ReplyCount        int      `json:"replyCount"`
	LikeCount         int      `json:"likeCount"`
	QuoteCount        int      `json:"quoteCount"`
	ViewCount         int      `json:"viewCount"`
	CreatedAt         string   `json:"createdAt"`
	Lang              string   `json:"lang"`
	BookmarkCount     int      `json:"bookmarkCount"`
	IsReply           bool     `json:"isReply"`
	InReplyToID       string   `json:"inReplyToId"`
	ConversationID    string   `json:"conversationId"`
	DisplayTextRange  []int    `json:"displayTextRange"`
	InReplyToUserID   string   `json:"inReplyToUserId"`
	InReplyToUsername string   `json:"inReplyToUsername"`
	Author            Author   `json:"author"`
	Entities          Entities `json:"entities"`
	QuotedTweet       *Tweet   `json:"quoted_tweet,omitempty"`
	RetweetedTweet    *Tweet   `json:"retweeted_tweet,omitempty"`
	IsLimitedReply    bool     `json:"isLimitedReply"`
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
