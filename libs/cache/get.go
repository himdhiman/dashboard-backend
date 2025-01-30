package cache

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

// Get retrieves a value from the cache and deserialize it into the result type
func (c *CacheClient) Get(ctx context.Context, key string, result interface{}) error {
	start := time.Now()
	data, err := c.client.Get(ctx, c.buildKey(key)).Bytes()
	if err == redis.Nil {
		return NewCacheMissError(key)
	}
	if err != nil {
		c.logger.WithFields(c.getLogFields(key)).WithError(err).Error("Failed to get cache")
		return err
	}

	err = c.deserializeValue(data, result)
	if err != nil {
		return NewCacheInvalidError("failed to deserialize value")
	}

	c.logger.WithFields(c.getLogFields(key)).WithField("duration", time.Since(start)).Debug("Cache hit")
	return nil
}

// GetSet atomically sets a new value and returns the old value
func (c *CacheClient) GetSet(ctx context.Context, key string, value interface{}, result interface{}) error {
	start := time.Now()
	data, err := c.serializeValue(value)
	if err != nil {
		return NewCacheInvalidError("failed to serialize value")
	}

	oldData, err := c.client.GetSet(ctx, c.buildKey(key), data).Bytes()
	if err != nil && err != redis.Nil {
		c.logger.WithFields(c.getLogFields(key)).WithError(err).Error("Failed to get-set cache")
		return err
	}

	if err != redis.Nil {
		if err := c.deserializeValue(oldData, result); err != nil {
			return NewCacheInvalidError("failed to deserialize value")
		}
	}

	c.logger.WithFields(c.getLogFields(key)).
		WithField("duration", time.Since(start)).
		Debug("Cache get-set")
	return nil
}
