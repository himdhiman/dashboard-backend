package cache

import (
	"context"

	"github.com/himdhiman/dashboard-backend/libs/logger"
)

// Delete removes a key from the cache
func (c *CacheClient) Delete(
	ctx context.Context,
	key string,
) error {
	if key == "" {
		return NewCacheInvalidError("key cannot be empty")
	}
	fullKey := c.buildKey(key)

	// Log delete operation
	c.logger.WithFields(logger.Fields{
		"key": key,
	}).Debug("Deleting key from cache")

	// Delete key from Redis
	if err := c.client.Del(ctx, fullKey).Err(); err != nil {
		return NewCacheDeleteError(key, err)
	}

	return nil
}

// Exists checks if a key exists in the cache
func (c *CacheClient) Exists(
	ctx context.Context,
	key string,
) (bool, error) {
	if key == "" {
		return false, NewCacheInvalidError("key cannot be empty")
	}
	fullKey := c.buildKey(key)

	// Log exists operation
	c.logger.WithFields(logger.Fields{
		"key": key,
	}).Debug("Checking if key exists in cache")

	count, err := c.client.Exists(ctx, fullKey).Result()
	if err != nil {
		return false, NewCacheExistsError(key, err)
	}

	return count > 0, nil
}
