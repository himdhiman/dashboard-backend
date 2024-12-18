package cache

import (
	"context"
	"time"
)

// Cacher defines the interface for cache operations with logging
type Cacher interface {
	// Basic Operations
	Set(ctx context.Context, key string, value interface{}, expiration ...time.Duration) error
	Get(ctx context.Context, key string, result interface{}) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)

	// Advanced Operations
	Increment(ctx context.Context, key string) (int64, error)
	Decrement(ctx context.Context, key string) (int64, error)
	SetNX(ctx context.Context, key string, value interface{}, expiration ...time.Duration) (bool, error)
	GetSet(ctx context.Context, key string, value interface{}, result interface{}) error

	// Connection Management
	Ping(ctx context.Context) error
	Close() error
}

// CacheOption allows for optional configuration of cache behavior
type CacheOption func(*CacheClient)