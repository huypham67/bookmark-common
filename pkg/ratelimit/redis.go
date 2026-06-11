package ratelimit

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type redisStore struct {
	client *redis.Client
}

// NewRedis returns a rate limit Store backed by Redis.
func NewRedis(client *redis.Client) Store {
	return &redisStore{
		client: client,
	}
}

// IncrementCounter increments the counter for the given key, setting its expiration only when the key is new.
func (r redisStore) IncrementCounter(context context.Context, key string, exp time.Duration) {
	r.client.Incr(context, key)
	r.client.ExpireNX(context, key, exp)
}

// GetCounter returns the current counter value for the given key, or 0 if it does not exist.
func (r redisStore) GetCounter(context context.Context, key string) (int, error) {
	result, err := r.client.Get(context, key).Int()
	if err != nil {
		if err == redis.Nil {
			return 0, nil
		}
		return 0, err
	}
	return result, nil
}
