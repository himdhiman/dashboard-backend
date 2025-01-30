package cache

import (
	"context"
	"time"
)

// Set stores a value in the cache with optional expiration
func (c *CacheClient) Set(ctx context.Context, key string, value interface{}, expiration ...time.Duration) error {
	start := time.Now()
	data, err := c.serializeValue(value)
	if err != nil {
		return NewCacheInvalidError("failed to serialize value")
	}

	exp := c.defaultTimeout
	if len(expiration) > 0 {
		exp = expiration[0]
	}

	err = c.client.Set(ctx, c.buildKey(key), data, exp).Err()
	if err != nil {
		c.logger.WithFields(c.getLogFields(key)).WithError(err).Error("Failed to set cache")
		return err
	}

	c.logger.WithFields(c.getLogFields(key)).WithField("duration", time.Since(start)).Debug("Cache set")
	return nil
}

// SetNX sets a value only if the key does not exist
func (c *CacheClient) SetNX(ctx context.Context, key string, value interface{}, expiration ...time.Duration) (bool, error) {
	start := time.Now()
	data, err := c.serializeValue(value)
	if err != nil {
		return false, NewCacheInvalidError("failed to serialize value")
	}

	exp := c.defaultTimeout
	if len(expiration) > 0 {
		exp = expiration[0]
	}

	ok, err := c.client.SetNX(ctx, c.buildKey(key), data, exp).Result()
	if err != nil {
		c.logger.WithFields(c.getLogFields(key)).WithError(err).Error("Failed to set cache NX")
		return false, err
	}

	c.logger.WithFields(c.getLogFields(key)).
		WithField("success", ok).
		WithField("duration", time.Since(start)).
		Debug("Cache set NX")
	return ok, nil
}
