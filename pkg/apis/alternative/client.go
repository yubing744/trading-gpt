package alternative

import (
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

var log = logrus.WithField("api", "alternative")

type AlternativeClient struct {
	baseURL string
	client  *http.Client
}

func NewAlternativeClient(opts ...Option) *AlternativeClient {
	cfg := &Options{
		baseURL: "https://api.alternative.me",
		timeout: time.Second * 20,
		debug:   false,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	return &AlternativeClient{
		baseURL: cfg.baseURL,
		client: &http.Client{
			Timeout:   cfg.timeout,
			Transport: cfg.transport,
		},
	}
}
