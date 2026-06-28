package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/huypham67/bookmark-common/pkg/jwt"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/rs/zerolog/log"
)

const (
	AuthorizationHeader = "Authorization"
	BearerScheme        = "Bearer"
	ClaimsKey           = "claims"

	eventAuthFailure = "AuthFailure"
	metricJWTFailure = "Custom/Auth/JWTFailure"
	reasonMissingHdr = "missing_header"
	reasonBadFormat  = "bad_format"
	reasonInvalidTkn = "invalid_token"
)

func recordAuthFailure(c *gin.Context, reason string) {
	app := newrelic.FromContext(c).Application()
	app.RecordCustomMetric(metricJWTFailure, 1)
	app.RecordCustomEvent(eventAuthFailure, map[string]interface{}{
		"reason":   reason,
		"path":     c.FullPath(),
		"clientIP": c.ClientIP(),
	})
}

// JWTAuth returns a Gin middleware function that validates JWT tokens in the Authorization header.
func JWTAuth(validator jwt.TokenValidator) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get Authorization header
		authHeader := c.GetHeader(AuthorizationHeader)
		if authHeader == "" {
			log.Warn().Msg("missing authorization header")
			recordAuthFailure(c, reasonMissingHdr)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "missing authorization header",
			})
			return
		}

		// Parse "Bearer <token>" format
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != BearerScheme {
			log.Warn().
				Str("auth_header", authHeader).
				Msg("invalid authorization header format")
			recordAuthFailure(c, reasonBadFormat)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid authorization header format",
			})
			return
		}

		tokenString := parts[1]

		// Validate token and extract claims
		claims, err := validator.ValidateToken(tokenString)
		if err != nil {
			log.Warn().
				Err(err).
				Msg("token validation failed")
			recordAuthFailure(c, reasonInvalidTkn)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid token",
			})
			return
		}

		// Store claims in context for handlers to access
		c.Set(ClaimsKey, claims)
		log.Debug().
			Str("user_id", claims.UserID).
			Msg("jwt token validated successfully")

		c.Next()
	}
}
