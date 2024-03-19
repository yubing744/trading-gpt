package anthropic

const (
	tokenEnvVarName = "ANTHROPIC_API_KEY" //nolint:gosec
)

type options struct {
	token   string
	baseURL string
}

type Option func(*options)

// WithToken passes the Anthropic API token to the client. If not set, the token
// is read from the ANTHROPIC_API_KEY environment variable.
func WithToken(token string) Option {
	return func(opts *options) {
		opts.token = token
	}
}

// WithBaseUrl passes the Anthropic base URL to the client.
// If not set, the default base URL is used.
func WithBaseURL(baseURL string) Option {
	return func(opts *options) {
		opts.baseURL = baseURL
	}
}
