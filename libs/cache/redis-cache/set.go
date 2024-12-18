package rediscache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"github.com/sirupsen/logrus"
)

// Set stores a value in the cache with optional expiration
func (c *CacheClient) Set(
	ctx context.Context, 
	key string, 
	value interface{}, 
	expiration ...time.Duration,
) error {
	// Determine expiration
	exp := c.defaultTimeout
	if len(expiration) > 0 {
		exp = expiration[0]
	}

	// Prepare logging fields
	logEntry := c.getLogFields(key).WithFields(logrus.Fields{
		"expiration": exp,
	})

	// Serialize the value
	var data []byte
	var err error
	switch v := value.(type) {
	case string:
		data = []byte(v)
		logEntry = logEntry.WithField("type", "string")
	case []byte:
		data = v
		logEntry = logEntry.WithField("type", "[]byte")
	default:
		data, err = json.Marshal(v)
		if err != nil {
			logEntry.WithError(err).Error("Failed to serialize value")
			return fmt.Errorf("failed to serialize value: %v", err)
		}
		logEntry = logEntry.WithField("type", "json")
	}

	// Set the key with full key path
	fullKey := c.buildKey(key)
	start := time.Now()
	err = c.client.Set(ctx, fullKey, data, exp).Err()
	
	if err != nil {
		logEntry.WithFields(logrus.Fields{
			"error":    err,
			"duration": time.Since(start),
		}).Error("Failed to set cache value")
		return err
	}

	logEntry.WithField("duration", time.Since(start)).Debug("Cache value set successfully")
	return nil
}

// SetNX sets a value only if the key does not exist
func (c *CacheClient) SetNX(
	ctx context.Context, 
	key string, 
	value interface{}, 
	expiration ...time.Duration,
) (bool, error) {
	// Determine expiration
	exp := c.defaultTimeout
	if len(expiration) > 0 {
		exp = expiration[0]
	}

	// Prepare logging fields
	logEntry := c.getLogFields(key).WithFields(logrus.Fields{
		"expiration": exp,
	})

	// Serialize the value
	var data []byte
	var err error
	switch v := value.(type) {
	case string:
		data = []byte(v)
		logEntry = logEntry.WithField("type", "string")
	case []byte:
		data = v
		logEntry = logEntry.WithField("type", "[]byte")
	default:
		data, err = json.Marshal(v)
		if err != nil {
			logEntry.WithError(err).Error("Failed to serialize value")
			return false, fmt.Errorf("failed to serialize value: %v", err)
		}
		logEntry = logEntry.WithField("type", "json")
	}

	fullKey := c.buildKey(key)
	start := time.Now()
	result, err := c.client.SetNX(ctx, fullKey, data, exp).Result()
	
	if err != nil {
		logEntry.WithFields(logrus.Fields{
			"error":    err,
			"duration": time.Since(start),
		}).Error("Failed to set NX cache value")
		return false, err
	}

	logEntry.WithFields(logrus.Fields{
		"set":      result,
		"duration": time.Since(start),
	}).Debug("SetNX operation completed")
	return result, nil
}