package flagcel

import (
	"net/http"
	"time"
)

const defaultPollInterval = 30 * time.Second

type providerConfig struct {
	httpClient   *http.Client
	pollInterval time.Duration
}

// Option configures a Flagcel OpenFeature provider.
type Option func(*providerConfig)

// WithHTTPClient sets the HTTP client used to fetch evaluation definitions.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(cfg *providerConfig) {
		cfg.httpClient = httpClient
	}
}

// WithPollInterval sets how often the provider polls /eval/definitions.
func WithPollInterval(interval time.Duration) Option {
	return func(cfg *providerConfig) {
		cfg.pollInterval = interval
	}
}

func newProviderConfig(options []Option) providerConfig {
	cfg := providerConfig{
		pollInterval: defaultPollInterval,
	}
	for _, option := range options {
		option(&cfg)
	}
	if cfg.pollInterval <= 0 {
		cfg.pollInterval = defaultPollInterval
	}
	return cfg
}
