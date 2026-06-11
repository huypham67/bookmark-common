package provider

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

// Config holds the rate limiting policy: at most Limit requests per Window for each caller.
type Config struct {
	Limit  int           `envconfig:"RATELIMIT_LIMIT" default:"20"`
	Window time.Duration `envconfig:"RATELIMIT_WINDOW" default:"10s"`
}

// LoadConfig loads rate limit configuration from environment variables with the given prefix.
func LoadConfig(prefix string) (*Config, error) {
	cfg := &Config{}
	if err := envconfig.Process(prefix, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
