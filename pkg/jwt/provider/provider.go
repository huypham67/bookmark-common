package provider

import (
	"fmt"

	"github.com/huypham67/bookmark-common/pkg/jwt"
)

// NewIssuer initializes a JWT token generator from environment configuration.
// The issuer signs tokens with the RSA private key, so only the token authority
// (e.g. user-service) should construct one.
func NewIssuer(envPrefix string) (jwt.TokenGenerator, error) {
	cfg, err := LoadIssuerConfig(envPrefix)
	if err != nil {
		return nil, fmt.Errorf("failed to load jwt config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid jwt config: %w", err)
	}

	privateKey, err := LoadRSAPrivateKeyFromFile(cfg.PrivateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load private key: %w", err)
	}

	generator, err := jwt.NewTokenGenerator(privateKey, cfg.Issuer, cfg.Audience, cfg.ExpirationDuration())
	if err != nil {
		return nil, fmt.Errorf("failed to create token generator: %w", err)
	}

	return generator, nil
}

// NewValidator initializes a JWT token validator from environment configuration.
// It loads ONLY the RSA public key, so relying parties (resource servers such as
// bookmark-service) never need access to the signing private key.
func NewValidator(envPrefix string) (jwt.TokenValidator, error) {
	cfg, err := LoadVerifierConfig(envPrefix)
	if err != nil {
		return nil, fmt.Errorf("failed to load jwt config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid jwt config: %w", err)
	}

	publicKey, err := LoadRSAPublicKeyFromFile(cfg.PublicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load public key: %w", err)
	}

	validator, err := jwt.NewTokenValidator(publicKey, cfg.Issuer, cfg.Audience)
	if err != nil {
		return nil, fmt.Errorf("failed to create token validator: %w", err)
	}

	return validator, nil
}
