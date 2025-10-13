package twitterapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Client implements the ITwitterClient interface
type Client struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
}

// NewClient creates a new Twitter API client with options
func NewClient(baseURL, apiKey string, opts ...ClientOption) ITwitterClient {
	c := &Client{
		BaseURL: baseURL,
		APIKey:  apiKey,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second, // default timeout
		},
	}

	// Apply options
	for _, opt := range opts {
		opt(c)
	}

	return c
}

// SearchTweets searches for tweets based on the query and returns the results
func (c *Client) SearchTweets(ctx context.Context, req *SearchRequest) (*SearchResponse, error) {
	// Build URL with query parameters
	apiURL := fmt.Sprintf("%s/twitter/tweet/advanced_search", c.BaseURL)
	u, err := url.Parse(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	// Add query parameters
	q := u.Query()
	q.Set("query", req.Query)

	if req.QueryType != "" {
		q.Set("queryType", req.QueryType)
	} else {
		q.Set("queryType", "Latest") // default to Latest
	}

	if req.Cursor != "" {
		q.Set("cursor", req.Cursor)
	}

	u.RawQuery = q.Encode()

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("X-API-Key", c.APIKey)
	httpReq.Header.Set("Accept", "application/json")

	// Execute request
	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Handle non-200 status codes
	if resp.StatusCode != http.StatusOK {
		message := "unknown error"
		switch resp.StatusCode {
		case http.StatusUnauthorized:
			message = "unauthorized: invalid API key"
		case http.StatusForbidden:
			message = "forbidden: insufficient permissions"
		case http.StatusTooManyRequests:
			message = "rate limit exceeded"
		case http.StatusGatewayTimeout:
			message = "request timed out"
		case http.StatusInternalServerError:
			message = "internal server error"
		}

		return nil, &HTTPError{
			StatusCode: resp.StatusCode,
			Message:    message,
			Body:       string(respBody),
		}
	}

	// Parse response
	var response SearchResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w, body: %s", err, string(respBody))
	}

	return &response, nil
}
