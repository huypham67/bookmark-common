package ratelimit

import (
	"context"
	"time"
)

// Limiter decides whether a caller identified by key may proceed under the configured policy.
//
//go:generate mockery --name=Limiter --output=./mocks --outpkg=mocks --filename=mock_limiter.go
type Limiter interface {
	Allow(ctx context.Context, key string) (bool, error)
}

type fixedWindowLimiter struct {
	store  Store
	limit  int
	window time.Duration
}

// NewFixedWindow creates a fixed-window Limiter that allows at most limit requests per window for each key.
func NewFixedWindow(store Store, limit int, window time.Duration) Limiter {
	return &fixedWindowLimiter{
		store:  store,
		limit:  limit,
		window: window,
	}
}

// Allow reports whether the request for key is within the limit, incrementing the counter when it is.
func (l *fixedWindowLimiter) Allow(ctx context.Context, key string) (bool, error) {
	count, err := l.store.GetCounter(ctx, key)
	if err != nil {
		return false, err
	}

	if count >= l.limit {
		return false, nil
	}

	l.store.IncrementCounter(ctx, key, l.window)
	return true, nil
}
