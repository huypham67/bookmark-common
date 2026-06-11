package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/huypham67/bookmark-common/pkg/jwt"
	"github.com/huypham67/bookmark-common/pkg/ratelimit/mocks"
	"github.com/stretchr/testify/assert"
)

func TestRateLimit(t *testing.T) {
	t.Parallel()

	const userID = "user-123"

	type expected struct {
		statusCode int
		nextCalled bool
	}

	testCases := []struct {
		name      string
		userID    string // when set, claims are injected so the key is by user
		setupMock func(context.Context, *mocks.Limiter)
		expected  expected
	}{
		{
			name:   "should allow when limiter permits the request",
			userID: userID,
			setupMock: func(ctx context.Context, l *mocks.Limiter) {
				l.On("Allow", ctx, "ratelimit:user:"+userID).Return(true, nil).Once()
			},
			expected: expected{statusCode: http.StatusOK, nextCalled: true},
		},
		{
			name:   "should block with 429 when limiter denies the request",
			userID: userID,
			setupMock: func(ctx context.Context, l *mocks.Limiter) {
				l.On("Allow", ctx, "ratelimit:user:"+userID).Return(false, nil).Once()
			},
			expected: expected{statusCode: http.StatusTooManyRequests, nextCalled: false},
		},
		{
			name:   "should fail open when limiter returns an error",
			userID: userID,
			setupMock: func(ctx context.Context, l *mocks.Limiter) {
				l.On("Allow", ctx, "ratelimit:user:"+userID).Return(false, errors.New("redis down")).Once()
			},
			expected: expected{statusCode: http.StatusOK, nextCalled: true},
		},
		{
			name:   "should key by client IP when unauthenticated",
			userID: "",
			setupMock: func(ctx context.Context, l *mocks.Limiter) {
				l.On("Allow", ctx, "ratelimit:ip:192.0.2.1").Return(true, nil).Once()
			},
			expected: expected{statusCode: http.StatusOK, nextCalled: true},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gin.SetMode(gin.TestMode)

			limiter := mocks.NewLimiter(t)

			nextCalled := false
			router := gin.New()
			router.GET("/resource",
				func(c *gin.Context) {
					if tc.userID != "" {
						c.Set("claims", &jwt.CustomClaims{UserID: tc.userID})
					}
					tc.setupMock(c, limiter)
				},
				RateLimit(limiter),
				func(c *gin.Context) {
					nextCalled = true
					c.Status(http.StatusOK)
				},
			)

			recorder := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/resource", nil)

			router.ServeHTTP(recorder, req)

			assert.Equal(t, tc.expected.statusCode, recorder.Code)
			assert.Equal(t, tc.expected.nextCalled, nextCalled)
		})
	}
}
