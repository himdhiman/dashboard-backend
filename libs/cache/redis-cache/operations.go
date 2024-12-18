package rediscache

import (
	"context"
)

// Increment increments a numeric key
func (c *CacheClient) Increment(
	ctx context.Context, 
	key string,
) (int64, error) {
	fullKey := c.buildKey(key)
	return c.client.Incr(ctx, fullKey).Result()
}

// Decrement decrements a numeric key
func (c *CacheClient) Decrement(
	ctx context.Context, 
	key string,
) (int64, error) {
	fullKey := c.buildKey(key)
	return c.client.Decr(ctx, fullKey).Result()
}