package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/huypham67/bookmark-common/pkg/jwt"
	jwtMocks "github.com/huypham67/bookmark-common/pkg/jwt/mocks"
	"github.com/stretchr/testify/assert"
)

func TestJWTAuth(t *testing.T) {
	t.Parallel()

	const (
		validToken = "valid-token"
		userID     = "user-123"
	)

	type expected struct {
		statusCode   int
		bodyContains string
		nextCalled   bool
	}

	testCases := []struct {
		name       string
		authHeader string
		setupMock  func(*jwtMocks.TokenValidator)
		expected   expected
	}{
		{
			name:       "should return 401 when authorization header is missing",
			authHeader: "",
			setupMock:  nil,
			expected: expected{
				statusCode:   http.StatusUnauthorized,
				bodyContains: "missing authorization header",
				nextCalled:   false,
			},
		},
		{
			name:       "should return 401 when header has no scheme",
			authHeader: validToken,
			setupMock:  nil,
			expected: expected{
				statusCode:   http.StatusUnauthorized,
				bodyContains: "invalid authorization header format",
				nextCalled:   false,
			},
		},
		{
			name:       "should return 401 when scheme is not Bearer",
			authHeader: "Basic " + validToken,
			setupMock:  nil,
			expected: expected{
				statusCode:   http.StatusUnauthorized,
				bodyContains: "invalid authorization header format",
				nextCalled:   false,
			},
		},
		{
			name:       "should return 401 when token is invalid",
			authHeader: "Bearer bad-token",
			setupMock: func(v *jwtMocks.TokenValidator) {
				v.On("ValidateToken", "bad-token").
					Return(nil, errors.New("token expired")).
					Once()
			},
			expected: expected{
				statusCode:   http.StatusUnauthorized,
				bodyContains: "invalid token",
				nextCalled:   false,
			},
		},
		{
			name:       "should call next and set claims when token is valid",
			authHeader: "Bearer " + validToken,
			setupMock: func(v *jwtMocks.TokenValidator) {
				v.On("ValidateToken", validToken).
					Return(&jwt.CustomClaims{UserID: userID}, nil).
					Once()
			},
			expected: expected{
				statusCode:   http.StatusOK,
				bodyContains: userID,
				nextCalled:   true,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gin.SetMode(gin.TestMode)

			validator := jwtMocks.NewTokenValidator(t)
			if tc.setupMock != nil {
				tc.setupMock(validator)
			}

			nextCalled := false
			router := gin.New()
			router.GET("/protected", JWTAuth(validator), func(c *gin.Context) {
				nextCalled = true
				// Verify the middleware stored the claims for downstream handlers.
				claims, err := jwt.GetClaims(c)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusOK, gin.H{"user_id": claims.UserID})
			})

			recorder := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			if tc.authHeader != "" {
				req.Header.Set(AuthorizationHeader, tc.authHeader)
			}

			router.ServeHTTP(recorder, req)

			assert.Equal(t, tc.expected.statusCode, recorder.Code)
			assert.Contains(t, recorder.Body.String(), tc.expected.bodyContains)
			assert.Equal(t, tc.expected.nextCalled, nextCalled)
		})
	}
}
