package cache

import (
	"context"
	"time"
)

// Increment increments a numeric key
func (c *CacheClient) Increment(ctx context.Context, key string) (int64, error) {
	start := time.Now()
	val, err := c.client.Incr(ctx, c.buildKey(key)).Result()
	if err != nil {
		c.logger.WithFields(c.getLogFields(key)).WithError(err).Error("Failed to increment cache")
		return 0, err
	}

	c.logger.WithFields(c.getLogFields(key)).
		WithField("value", val).
		WithField("duration", time.Since(start)).
		Debug("Cache incremented")
	return val, nil
}

// Decrement decrements a numeric key
func (c *CacheClient) Decrement(ctx context.Context, key string) (int64, error) {
	start := time.Now()
	val, err := c.client.Decr(ctx, c.buildKey(key)).Result()
	if err != nil {
		c.logger.WithFields(c.getLogFields(key)).WithError(err).Error("Failed to decrement cache")
		return 0, err
	}

	c.logger.WithFields(c.getLogFields(key)).
		WithField("value", val).
		WithField("duration", time.Since(start)).
		Debug("Cache decremented")
	return val, nil
}
