package alternative

import (
	"net/http"
	"time"
)

type Options struct {
	baseURL   string
	timeout   time.Duration
	transport http.RoundTripper
	debug     bool
}

type Option func(opts *Options)

func WithOptions(options Options) Option {
	return func(opts *Options) {
		*opts = options
	}
}

func WithBaseURL(baseURL string) Option {
	return func(opts *Options) {
		opts.baseURL = baseURL
	}
}

func WithTimeout(timeout time.Duration) Option {
	return func(opts *Options) {
		opts.timeout = timeout
	}
}

func WithTransport(transport http.RoundTripper) Option {
	return func(opts *Options) {
		opts.transport = transport
	}
}

func WithDebug(debug bool) Option {
	return func(opts *Options) {
		opts.debug = debug
	}
}
