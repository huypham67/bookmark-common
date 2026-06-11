package redis

import (
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

// Mock provides a mock Redis server and client for testing purposes.
type Mock struct {
	Server *miniredis.Miniredis
	Client *redis.Client
}

// NewMock creates a new mock Redis server and client for testing purposes.
func NewMock(t *testing.T) *Mock {
	t.Helper()

	server, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to start mock Redis server: %v", err)
	}

	t.Cleanup(func() {
		server.Close()
	})

	client := redis.NewClient(&redis.Options{
		Addr: server.Addr(),
	})

	return &Mock{
		Server: server,
		Client: client,
	}
}

// FastForward advances the mock Redis server's internal clock by the specified duration, allowing tests to simulate time-based behavior such as key expiration.
func (m *Mock) FastForward(duration time.Duration) {
	m.Server.FastForward(duration)
}

// Close shuts down the mock Redis server and closes the client connection.
func (m *Mock) Close() {
	err := m.Client.Close()
	if err != nil {
		return
	}
	m.Server.Close()
}
