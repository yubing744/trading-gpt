package native

import (
	"net/http"

	"github.com/madebywelch/anthropic-go/v3/pkg/anthropic"
)

const (
	// AnthropicAPIVersion is the version of the Anthropics API that this client is compatible with.
	AnthropicAPIVersion = "2023-06-01"
)

// doRequest sends an HTTP request and returns the response, handling any non-OK HTTP status codes.
func (c *Client) doRequest(request *http.Request) (*http.Response, error) {
	request.Header.Add("anthropic-version", AnthropicAPIVersion)

	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		err = anthropic.MapHTTPStatusCodeToError(response.StatusCode)
		return nil, err
	}

	return response, nil
}
