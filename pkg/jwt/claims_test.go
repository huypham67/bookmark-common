package jwt

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestContext() *gin.Context {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	return c
}

func TestGetClaims(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		setup  func(*gin.Context)
		verify func(*testing.T, *CustomClaims, error)
	}{
		{
			name: "should return claims when present in context",
			setup: func(c *gin.Context) {
				c.Set("claims", &CustomClaims{UserID: "user-123", Email: "john@example.com"})
			},
			verify: func(t *testing.T, claims *CustomClaims, err error) {
				require.NoError(t, err)
				require.NotNil(t, claims)
				assert.Equal(t, "user-123", claims.UserID)
				assert.Equal(t, "john@example.com", claims.Email)
			},
		},
		{
			name:  "should return ErrMissingClaims when claims absent",
			setup: func(c *gin.Context) {},
			verify: func(t *testing.T, claims *CustomClaims, err error) {
				require.ErrorIs(t, err, ErrMissingClaims)
				assert.Nil(t, claims)
			},
		},
		{
			name: "should return ErrInvalidClaims when claims have wrong type",
			setup: func(c *gin.Context) {
				c.Set("claims", "not-a-claims-struct")
			},
			verify: func(t *testing.T, claims *CustomClaims, err error) {
				require.ErrorIs(t, err, ErrInvalidClaims)
				assert.Nil(t, claims)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			c := newTestContext()
			tc.setup(c)

			claims, err := GetClaims(c)

			tc.verify(t, claims, err)
		})
	}
}

func TestGetUserIDFromContext(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		setup  func(*gin.Context)
		verify func(*testing.T, string, error)
	}{
		{
			name: "should return user ID when claims present",
			setup: func(c *gin.Context) {
				c.Set("claims", &CustomClaims{UserID: "user-123"})
			},
			verify: func(t *testing.T, userID string, err error) {
				require.NoError(t, err)
				assert.Equal(t, "user-123", userID)
			},
		},
		{
			name:  "should propagate error when claims absent",
			setup: func(c *gin.Context) {},
			verify: func(t *testing.T, userID string, err error) {
				require.ErrorIs(t, err, ErrMissingClaims)
				assert.Empty(t, userID)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			c := newTestContext()
			tc.setup(c)

			userID, err := GetUserIDFromContext(c)

			tc.verify(t, userID, err)
		})
	}
}
