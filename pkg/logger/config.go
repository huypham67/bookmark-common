package logger

import "github.com/kelseyhightower/envconfig"

// Config holds the configuration for the logger.
type Config struct {
	Level string `envconfig:"LOG_LEVEL" default:"info"`
}

// LoadConfig loads logger configuration from environment variables with the given prefix.
func LoadConfig(prefix string) (*Config, error) {
	cfg := &Config{}

	if err := envconfig.Process(prefix, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
