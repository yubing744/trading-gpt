package native

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/madebywelch/anthropic-go/v3/pkg/anthropic"
)

type Client struct {
	httpClient *http.Client
	apiKey     string
	baseURL    string
	beta       string
}

type Config struct {
	APIKey  string
	BaseURL string
	Beta    string

	// Optional (defaults to http.DefaultClient)
	HTTPClient *http.Client
}

func MakeClient(cfg Config) (*Client, error) {
	if cfg.APIKey == "" {
		return nil, anthropic.ErrAnthropicApiKeyRequired
	}

	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.anthropic.com"
	}

	if cfg.HTTPClient == nil {
		cfg.HTTPClient = http.DefaultClient
	}

	return &Client{
		httpClient: cfg.HTTPClient,
		apiKey:     cfg.APIKey,
		baseURL:    cfg.BaseURL,
		beta:       cfg.Beta,
	}, nil
}

func (c *Client) Message(ctx context.Context, req *anthropic.MessageRequest) (*anthropic.MessageResponse, error) {
	err := ValidateMessageRequest(req)
	if err != nil {
		return nil, err
	}

	return c.sendMessageRequest(ctx, req)
}

func (c *Client) sendMessageRequest(
	ctx context.Context,
	req *anthropic.MessageRequest,
) (*anthropic.MessageResponse, error) {
	// Marshal the request to JSON
	data, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("error marshalling message request: %w", err)
	}

	// Create the HTTP request
	requestURL := fmt.Sprintf("%s/v1/messages", c.baseURL)
	request, err := http.NewRequestWithContext(ctx, "POST", requestURL, bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("error creating new request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Api-Key", c.apiKey)
	if len(c.beta) > 0 {
		request.Header.Set("anthropic-beta", c.beta)
	}

	// Use the doRequest method to send the HTTP request
	response, err := c.doRequest(request)
	if err != nil {
		return nil, fmt.Errorf("error sending message request: %w", err)
	}
	defer response.Body.Close()

	// Decode the response body to a MessageResponse object
	messageResponse := &anthropic.MessageResponse{}
	err = json.NewDecoder(response.Body).Decode(messageResponse)
	if err != nil {
		return nil, fmt.Errorf("error decoding message response: %w", err)
	}

	return messageResponse, nil
}

func ValidateMessageRequest(req *anthropic.MessageRequest) error {
	if req.Stream {
		return fmt.Errorf("cannot use Message with streaming enabled, use MessageStream instead")
	}

	if !req.Model.IsImageCompatible() && req.ContainsImageContent() {
		return fmt.Errorf("model %s does not support image content", req.Model)
	}

	if req.CountImageContent() > 20 {
		return fmt.Errorf("too many image content blocks, maximum is 20")
	}

	return nil
}
