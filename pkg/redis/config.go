package redis

import "github.com/kelseyhightower/envconfig"

// Config holds the Redis connection configuration.
type Config struct {
	Addr     string `envconfig:"REDIS_ADDR" default:"localhost:6379"`
	Password string `envconfig:"REDIS_PASSWORD" default:""`
	Database int    `envconfig:"REDIS_DATABASE" default:"0"`
}

// LoadConfig loads Redis configuration from environment variables with the given prefix.
func LoadConfig(prefix string) (*Config, error) {
	cfg := &Config{}
	err := envconfig.Process(prefix, cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
