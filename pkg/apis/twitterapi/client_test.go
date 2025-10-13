package twitterapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	client := NewClient("https://api.twitterapi.io", "test-api-key")
	if client == nil {
		t.Fatal("Expected client to be created")
	}

	c, ok := client.(*Client)
	if !ok {
		t.Fatal("Expected client to be of type *Client")
	}

	if c.BaseURL != "https://api.twitterapi.io" {
		t.Errorf("Expected BaseURL to be 'https://api.twitterapi.io', got '%s'", c.BaseURL)
	}

	if c.APIKey != "test-api-key" {
		t.Errorf("Expected APIKey to be 'test-api-key', got '%s'", c.APIKey)
	}
}

func TestSearchTweets_Success(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify headers
		if r.Header.Get("X-API-Key") != "test-api-key" {
			t.Errorf("Expected X-API-Key header to be 'test-api-key', got '%s'", r.Header.Get("X-API-Key"))
		}

		// Verify query parameters
		query := r.URL.Query().Get("query")
		if query != "bitcoin" {
			t.Errorf("Expected query parameter to be 'bitcoin', got '%s'", query)
		}

		// Return mock response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"tweets": [
				{
					"id": "123",
					"text": "Test tweet about bitcoin",
					"author": {
						"username": "testuser",
						"name": "Test User"
					},
					"like_count": 10,
					"retweet_count": 5,
					"created_at": "2025-01-01T00:00:00Z"
				}
			],
			"has_next_page": false,
			"next_cursor": ""
		}`))
	}))
	defer server.Close()

	// Create client with mock server URL
	client := NewClient(server.URL, "test-api-key", WithTimeout(5*time.Second))

	// Test search
	ctx := context.Background()
	req := &SearchRequest{
		Query:     "bitcoin",
		QueryType: "Latest",
	}

	resp, err := client.SearchTweets(ctx, req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(resp.Tweets) != 1 {
		t.Errorf("Expected 1 tweet, got %d", len(resp.Tweets))
	}

	if resp.Tweets[0].Text != "Test tweet about bitcoin" {
		t.Errorf("Expected tweet text to be 'Test tweet about bitcoin', got '%s'", resp.Tweets[0].Text)
	}
}

func TestSearchTweets_Unauthorized(t *testing.T) {
	// Create a mock server that returns 401
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "Unauthorized"}`))
	}))
	defer server.Close()

	// Create client with mock server URL
	client := NewClient(server.URL, "invalid-key", WithTimeout(5*time.Second))

	// Test search
	ctx := context.Background()
	req := &SearchRequest{
		Query:     "bitcoin",
		QueryType: "Latest",
	}

	_, err := client.SearchTweets(ctx, req)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	httpErr, ok := err.(*HTTPError)
	if !ok {
		t.Fatalf("Expected HTTPError, got %T", err)
	}

	if httpErr.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status code 401, got %d", httpErr.StatusCode)
	}
}

func TestSearchTweets_RateLimit(t *testing.T) {
	// Create a mock server that returns 429
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{"error": "Rate limit exceeded"}`))
	}))
	defer server.Close()

	// Create client with mock server URL
	client := NewClient(server.URL, "test-api-key", WithTimeout(5*time.Second))

	// Test search
	ctx := context.Background()
	req := &SearchRequest{
		Query:     "bitcoin",
		QueryType: "Latest",
	}

	_, err := client.SearchTweets(ctx, req)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	httpErr, ok := err.(*HTTPError)
	if !ok {
		t.Fatalf("Expected HTTPError, got %T", err)
	}

	if httpErr.StatusCode != http.StatusTooManyRequests {
		t.Errorf("Expected status code 429, got %d", httpErr.StatusCode)
	}

	if httpErr.Message != "rate limit exceeded" {
		t.Errorf("Expected message 'rate limit exceeded', got '%s'", httpErr.Message)
	}
}
