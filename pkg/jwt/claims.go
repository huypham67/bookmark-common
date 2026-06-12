package jwt

import (
	"errors"
	"time"

	"github.com/gin-gonic/gin"
	gojwt "github.com/golang-jwt/jwt/v5"
)

// CustomClaims represents JWT custom claims containing user information and standard registered claims.
type CustomClaims struct {
	UserID      string `json:"id"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
	gojwt.RegisteredClaims
}

func newCustomClaims(userID, displayName, email string, expiry time.Duration, issuer, audience string) *CustomClaims {
	return &CustomClaims{
		UserID:      userID,
		DisplayName: displayName,
		Email:       email,
		RegisteredClaims: gojwt.RegisteredClaims{
			ExpiresAt: gojwt.NewNumericDate(time.Now().Add(expiry)),
			IssuedAt:  gojwt.NewNumericDate(time.Now()),
			Issuer:    issuer,
			Audience:  gojwt.ClaimStrings{audience},
		},
	}
}

var (
	ErrMissingClaims = errors.New("missing claims in context")
	ErrInvalidClaims = errors.New("invalid claims type")
)

// GetClaims retrieves the CustomClaims from the Gin context set by the JWTAuth middleware.
func GetClaims(c *gin.Context) (*CustomClaims, error) {
	claimsObj, exists := c.Get("claims")
	if !exists {
		return nil, ErrMissingClaims
	}

	claims, ok := claimsObj.(*CustomClaims)
	if !ok {
		return nil, ErrInvalidClaims
	}

	return claims, nil
}

// GetUserIDFromContext extracts the user ID from the JWT claims in the Gin context.
func GetUserIDFromContext(c *gin.Context) (string, error) {
	claims, err := GetClaims(c)
	if err != nil {
		return "", err
	}

	return claims.UserID, nil
}
