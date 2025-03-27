package coze

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client implements the ICozeClient interface
type Client struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
}

// NewClient creates a new Coze API client with options
func NewClient(baseURL, apiKey string, opts ...ClientOption) ICozeClient {
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

// Chat sends a chat request to the Coze API and handles the response
func (c *Client) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	url := fmt.Sprintf("%s/open_api/v2/chat", c.BaseURL)
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.APIKey))

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var response ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil
}

// RunWorkflow executes a workflow and returns the result
func (c *Client) RunWorkflow(ctx context.Context, req *WorkflowRequest) (*WorkflowResponse, error) {
	url := fmt.Sprintf("%s/v1/workflow/run", c.BaseURL)
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.APIKey))

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
			message = "workflow execution timed out"
		case http.StatusInternalServerError:
			message = "internal server error"
		}

		return nil, &HTTPError{
			StatusCode: resp.StatusCode,
			Message:    message,
			Body:       string(respBody),
		}
	}

	var response WorkflowResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Handle API error responses
	if response.Code != 0 {
		return nil, &CozeError{
			Code:    response.Code,
			Message: response.Msg,
		}
	}

	return &response, nil
}
