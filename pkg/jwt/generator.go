package jwt

import (
	"crypto/rsa"
	"time"

	gojwt "github.com/golang-jwt/jwt/v5"
)

// TokenGenerator defines the contract for JWT token generation operations.
//
//go:generate mockery --name=TokenGenerator --output=./mocks --outpkg=mocks --filename=generator.go
type TokenGenerator interface {
	GenerateToken(userID, displayName, email string) (string, error)
}

type rsaTokenGenerator struct {
	privateKey *rsa.PrivateKey
	issuer     string
	audience   string
	expiry     time.Duration
}

// NewTokenGenerator creates a new token generator with the given RSA private key, issuer, audience, and token expiry duration.
func NewTokenGenerator(privateKey *rsa.PrivateKey, issuer, audience string, expiry time.Duration) (TokenGenerator, error) {
	return &rsaTokenGenerator{
		privateKey: privateKey,
		issuer:     issuer,
		audience:   audience,
		expiry:     expiry,
	}, nil
}

// GenerateToken generates a JWT token string with the given user information and configured claims.
func (g *rsaTokenGenerator) GenerateToken(userID, displayName, email string) (string, error) {
	claims := newCustomClaims(userID, displayName, email, g.expiry, g.issuer, g.audience)
	token := gojwt.NewWithClaims(gojwt.SigningMethodRS256, claims)
	return token.SignedString(g.privateKey)
}
