package anthropic

const (
	tokenEnvVarName = "ANTHROPIC_API_KEY" //nolint:gosec
)

type options struct {
	token          string
	baseURL        string
	thinkingBudget int64
	enableThinking bool
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

// WithThinkingBudget sets the thinking budget in tokens for Claude's extended thinking.
// This enables extended thinking mode when budget > 0.
func WithThinkingBudget(budget int64) Option {
	return func(opts *options) {
		opts.thinkingBudget = budget
		opts.enableThinking = budget > 0
	}
}

// WithThinking enables or disables Claude's extended thinking mode.
// When enabled, uses the default thinking budget (1024 tokens).
func WithThinking(enabled bool) Option {
	return func(opts *options) {
		opts.enableThinking = enabled
		if enabled && opts.thinkingBudget == 0 {
			opts.thinkingBudget = 1024 // Default thinking budget
		}
	}
}
