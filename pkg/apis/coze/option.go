package coze

import (
	"net/http"
	"time"
)

// ClientOption defines the type for client options
type ClientOption func(*Client)

// WithTimeout sets custom timeout for HTTP client
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.HTTPClient.Timeout = timeout
	}
}

// WithTransport sets custom transport for HTTP client
func WithTransport(transport http.RoundTripper) ClientOption {
	return func(c *Client) {
		c.HTTPClient.Transport = transport
	}
}
