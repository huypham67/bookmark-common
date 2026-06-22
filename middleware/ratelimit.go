package middleware

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/huypham67/bookmark-common/pkg/jwt"
	"github.com/huypham67/bookmark-common/pkg/ratelimit"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/rs/zerolog/log"
)

const (
	rateLimitKeyPrefix      = "ratelimit:"
	eventRateLimitExceeded  = "RateLimitExceeded"
	metricRateLimitExceeded = "Custom/RateLimit/Exceeded"
)

// RateLimit returns a Gin middleware that enforces the limiter's policy per caller.
func RateLimit(limiter ratelimit.Limiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := fmt.Sprintf("%s%s", rateLimitKeyPrefix, rateLimitIdentifier(c))

		allowed, err := limiter.Allow(c, key)
		if err != nil {
			// Fail-open: a broken rate limiter must not take the whole service down.
			log.Error().Err(err).Msg("rate limit store error, allowing request")
			c.Next()
			return
		}

		if !allowed {
			log.Warn().Str("key", key).Msg("rate limit exceeded")
			app := newrelic.FromContext(c).Application()
			app.RecordCustomMetric(metricRateLimitExceeded, 1)
			app.RecordCustomEvent(eventRateLimitExceeded, map[string]interface{}{
				"identifier": rateLimitIdentifier(c),
				"path":       c.FullPath(),
			})
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded",
			})
			return
		}

		c.Next()
	}
}

// rateLimitIdentifier keys the limit by authenticated user, falling back to client IP for public callers.
func rateLimitIdentifier(c *gin.Context) string {
	if userID, err := jwt.GetUserIDFromContext(c); err == nil && userID != "" {
		return "user:" + userID
	}
	return "ip:" + c.ClientIP()
}
