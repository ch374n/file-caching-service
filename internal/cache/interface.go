package cache

import "context"

// Cache defines the interface for caching operations
// This allows for easy mocking in tests
type Cache interface {
	Get(ctx context.Context, key string) ([]byte, bool, error)
	Set(ctx context.Context, key string, data []byte) error
	Ping(ctx context.Context) error
	Close() error
}

// Ensure RedisCache implements Cache interface
var _ Cache = (*RedisCache)(nil)
