package ratelimit

import (
	"context"
	"testing"
	"time"

	"github.com/huypham67/bookmark-common/pkg/redis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testWindow = 10 * time.Second

func newTestStore(t *testing.T) (Store, *redis.Mock) {
	t.Helper()

	mockRedis := redis.NewMock(t)
	store := NewRedis(mockRedis.Client)

	return store, mockRedis
}

func TestRedisStore_IncrementCounter(t *testing.T) {
	t.Parallel()

	const key = "ratelimit:user:123"

	testCases := []struct {
		name   string
		setup  func(context.Context, Store, *redis.Mock)
		verify func(*testing.T, context.Context, Store, *redis.Mock)
	}{
		{
			name:  "should create counter with value 1 on first increment",
			setup: func(_ context.Context, _ Store, _ *redis.Mock) {},
			verify: func(t *testing.T, ctx context.Context, store Store, _ *redis.Mock) {
				count, err := store.GetCounter(ctx, key)
				require.NoError(t, err)
				assert.Equal(t, 1, count)
			},
		},
		{
			name: "should accumulate on repeated increments",
			setup: func(ctx context.Context, store Store, _ *redis.Mock) {
				store.IncrementCounter(ctx, key, testWindow)
				store.IncrementCounter(ctx, key, testWindow)
			},
			verify: func(t *testing.T, ctx context.Context, store Store, _ *redis.Mock) {
				count, err := store.GetCounter(ctx, key)
				require.NoError(t, err)
				assert.Equal(t, 3, count)
			},
		},
		{
			name:  "should set expiration on a new key",
			setup: func(_ context.Context, _ Store, _ *redis.Mock) {},
			verify: func(t *testing.T, _ context.Context, _ Store, mockRedis *redis.Mock) {
				assert.Equal(t, testWindow, mockRedis.Server.TTL(key))
			},
		},
		{
			name: "should not reset expiration on an existing key",
			setup: func(ctx context.Context, store Store, mockRedis *redis.Mock) {
				store.IncrementCounter(ctx, key, testWindow)
				mockRedis.FastForward(4 * time.Second)
			},
			verify: func(t *testing.T, _ context.Context, _ Store, mockRedis *redis.Mock) {
				// ExpireNX is a no-op on the existing key, so the TTL keeps counting down.
				assert.Equal(t, testWindow-4*time.Second, mockRedis.Server.TTL(key))
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			store, mockRedis := newTestStore(t)

			tc.setup(ctx, store, mockRedis)
			store.IncrementCounter(ctx, key, testWindow)

			tc.verify(t, ctx, store, mockRedis)
		})
	}
}

func TestRedisStore_GetCounter(t *testing.T) {
	t.Parallel()

	const key = "ratelimit:user:123"

	testCases := []struct {
		name      string
		setup     func(context.Context, Store)
		wantCount int
		wantErr   bool
	}{
		{
			name:      "should return 0 when key does not exist",
			setup:     func(_ context.Context, _ Store) {},
			wantCount: 0,
		},
		{
			name: "should return current value when key exists",
			setup: func(ctx context.Context, store Store) {
				store.IncrementCounter(ctx, key, testWindow)
				store.IncrementCounter(ctx, key, testWindow)
			},
			wantCount: 2,
		},
		{
			name:    "should return error when redis is unavailable",
			setup:   func(_ context.Context, _ Store) {},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			store, mockRedis := newTestStore(t)

			tc.setup(ctx, store)
			if tc.wantErr {
				mockRedis.Close()
			}

			count, err := store.GetCounter(ctx, key)

			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.wantCount, count)
		})
	}
}
