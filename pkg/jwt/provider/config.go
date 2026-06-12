package provider

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

// CommonConfig holds the JWT claim fields shared by both issuing and verifying.
// Embedding it in IssuerConfig and VerifierConfig keeps issuer/audience handling
// (and its validation) in one place.
type CommonConfig struct {
	Issuer   string `envconfig:"JWT_ISSUER" default:"bookmark-service"`
	Audience string `envconfig:"JWT_AUDIENCE" default:"bookmark-service"`
}

// Validate ensures the shared claim fields are present.
func (c *CommonConfig) Validate() error {
	if c.Issuer == "" {
		return fmt.Errorf("jwt config: issuer is required")
	}
	if c.Audience == "" {
		return fmt.Errorf("jwt config: audience is required")
	}
	return nil
}

// VerifierConfig holds the fields required to validate (verify) JWT tokens.
// A verifier needs only the public key — never the signing private key.
type VerifierConfig struct {
	CommonConfig
	PublicKeyPath string `envconfig:"JWT_PUBLIC_KEY_PATH" default:"keys/public.pem"`
}

// Validate ensures the fields required to verify tokens are present.
func (c *VerifierConfig) Validate() error {
	if c.PublicKeyPath == "" {
		return fmt.Errorf("jwt config: public key path is required")
	}
	return c.CommonConfig.Validate()
}

// IssuerConfig holds the fields required to issue (sign) JWT tokens.
type IssuerConfig struct {
	CommonConfig
	PrivateKeyPath    string `envconfig:"JWT_PRIVATE_KEY_PATH" default:"keys/private.pem"`
	ExpirationSeconds int64  `envconfig:"JWT_EXPIRATION_SECONDS" default:"3600"`
}

// Validate ensures the fields required to issue tokens are present.
func (c *IssuerConfig) Validate() error {
	if c.PrivateKeyPath == "" {
		return fmt.Errorf("jwt config: private key path is required")
	}
	if c.ExpirationSeconds <= 0 {
		return fmt.Errorf("jwt config: expiration seconds must be greater than 0")
	}
	return c.CommonConfig.Validate()
}

// ExpirationDuration returns the token expiration as a time.Duration.
func (c *IssuerConfig) ExpirationDuration() time.Duration {
	return time.Duration(c.ExpirationSeconds) * time.Second
}

// LoadIssuerConfig loads issuer JWT configuration from environment variables
// using the specified prefix.
func LoadIssuerConfig(prefix string) (*IssuerConfig, error) {
	cfg := &IssuerConfig{}
	if err := envconfig.Process(prefix, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

// LoadVerifierConfig loads verifier JWT configuration from environment variables
// using the specified prefix.
func LoadVerifierConfig(prefix string) (*VerifierConfig, error) {
	cfg := &VerifierConfig{}
	if err := envconfig.Process(prefix, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
