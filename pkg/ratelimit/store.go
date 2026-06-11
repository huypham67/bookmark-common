package ratelimit

import (
	"context"
	"time"
)

// Store defines the interface for rate limit storage backends (e.g., in-memory, Redis).
//
//go:generate mockery --name=Store --output=./mocks --outpkg=mocks --filename=mock_store.go
type Store interface {
	IncrementCounter(context context.Context, key string, exp time.Duration)
	GetCounter(context context.Context, key string) (int, error)
}
