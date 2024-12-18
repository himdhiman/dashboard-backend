package cache

import (
	"time"

	"github.com/go-redis/redis/v8"
)

// WithTimeout sets a custom default timeout for cache entries
func WithTimeout(timeout time.Duration) CacheOption {
	return func(c *CacheClient) {
		c.defaultTimeout = timeout
	}
}

// WithPrefix sets a custom prefix for cache keys
func WithPrefix(prefix string) CacheOption {
	return func(c *CacheClient) {
		c.prefix = prefix
	}
}

// WithCustomRedisClient allows passing a pre-configured Redis client
func WithCustomRedisClient(client *redis.Client) CacheOption {
	return func(c *CacheClient) {
		c.client = client
	}
}