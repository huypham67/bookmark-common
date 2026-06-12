package provider

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

// Config holds JWT-related configuration for token generation and validation.
type Config struct {
	PrivateKeyPath    string `envconfig:"JWT_PRIVATE_KEY_PATH" default:"keys/private.pem"`
	PublicKeyPath     string `envconfig:"JWT_PUBLIC_KEY_PATH" default:"keys/public.pem"`
	Issuer            string `envconfig:"JWT_ISSUER" default:"bookmark-service"`
	Audience          string `envconfig:"JWT_AUDIENCE" default:"bookmark-service"`
	ExpirationSeconds int64  `envconfig:"JWT_EXPIRATION_SECONDS" default:"3600"`
}

// LoadConfig loads JWT configuration from environment variables using the specified prefix.
func LoadConfig(prefix string) (*Config, error) {
	cfg := &Config{}
	err := envconfig.Process(prefix, cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

// Validate ensures that all required JWT configuration fields are valid.
func (c *Config) Validate() error {
	if c.PrivateKeyPath == "" {
		return fmt.Errorf("jwt config: private key path is required")
	}
	if c.PublicKeyPath == "" {
		return fmt.Errorf("jwt config: public key path is required")
	}
	if c.Issuer == "" {
		return fmt.Errorf("jwt config: issuer is required")
	}
	if c.Audience == "" {
		return fmt.Errorf("jwt config: audience is required")
	}
	if c.ExpirationSeconds <= 0 {
		return fmt.Errorf("jwt config: expiration seconds must be greater than 0")
	}
	return nil
}

// ExpirationDuration returns the token expiration as a time.Duration.
func (c *Config) ExpirationDuration() time.Duration {
	return time.Duration(c.ExpirationSeconds) * time.Second
}
