package cache

import (
	"context"
)

// Delete removes a key from the cache
func (c *CacheClient) Delete(
	ctx context.Context, 
	key string,
) error {
	fullKey := c.buildKey(key)
	return c.client.Del(ctx, fullKey).Err()
}

// Exists checks if a key exists in the cache
func (c *CacheClient) Exists(
	ctx context.Context, 
	key string,
) (bool, error) {
	fullKey := c.buildKey(key)
	count, err := c.client.Exists(ctx, fullKey).Result()
	return count > 0, err
}