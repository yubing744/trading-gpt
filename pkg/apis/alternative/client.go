package alternative

import (
	"net/http"
	"time"
)

type AlternativeClient struct {
	client *http.Client
}

func NewAlternativeClient(opts ...Option) *AlternativeClient {
	cfg := &Options{
		baseURL: "https://alternative.me",
		timeout: time.Second * 20,
		debug:   false,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	return &AlternativeClient{
		client: &http.Client{
			Timeout:   cfg.timeout,
			Transport: cfg.transport,
		},
	}
}
