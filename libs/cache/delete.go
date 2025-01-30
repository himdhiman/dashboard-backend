package cache

import (
	"context"
	"time"
)

// Delete removes a key from the cache
func (c *CacheClient) Delete(ctx context.Context, key string) error {
	start := time.Now()
	err := c.client.Del(ctx, c.buildKey(key)).Err()
	if err != nil {
		c.logger.WithFields(c.getLogFields(key)).WithError(err).Error("Failed to delete cache")
		return err
	}

	c.logger.WithFields(c.getLogFields(key)).WithField("duration", time.Since(start)).Debug("Cache deleted")
	return nil
}

// Exists checks if a key exists in the cache
func (c *CacheClient) Exists(ctx context.Context, key string) (bool, error) {
	start := time.Now()
	n, err := c.client.Exists(ctx, c.buildKey(key)).Result()
	if err != nil {
		c.logger.WithFields(c.getLogFields(key)).WithError(err).Error("Failed to check cache existence")
		return false, err
	}

	exists := n > 0
	c.logger.WithFields(c.getLogFields(key)).
		WithField("exists", exists).
		WithField("duration", time.Since(start)).
		Debug("Cache existence checked")
	return exists, nil
}
