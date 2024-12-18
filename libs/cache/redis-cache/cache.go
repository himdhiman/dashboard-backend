package rediscache

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

// CacheClient implements the Cacher interface with logging
type CacheClient struct {
	client         *redis.Client
	defaultTimeout time.Duration
	prefix         string
	logger         *logrus.Logger
}

// NewCacheClient creates a new Redis cache client with optional configurations
func NewCacheClient(
	host string, 
	port int, 
	password string, 
	db int, 
	logger *logrus.Logger,
	options ...CacheOption,
) *CacheClient {
	// Validate logger
	if logger == nil {
		logger = logrus.New()
		logger.Warning("No logger provided, using default logger")
	}

	// Create Redis client configuration
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", host, port),
		Password: password,
		DB:       db,
	})

	// Create cache client
	cache := &CacheClient{
		client:         rdb,
		defaultTimeout: 1 * time.Hour, // Default timeout
		prefix:         "app:", // Default prefix
		logger:         logger,
	}

	// Apply optional configurations
	for _, opt := range options {
		opt(cache)
	}

	// Log cache client initialization
	cache.logger.WithFields(logrus.Fields{
		"host":     host,
		"port":     port,
		"database": db,
		"prefix":   cache.prefix,
	}).Info("Redis cache client initialized")

	return cache
}

// buildKey constructs the full cache key with prefix
func (c *CacheClient) buildKey(key string) string {
	return fmt.Sprintf("%s%s", c.prefix, key)
}

// Ping checks the connection to Redis
func (c *CacheClient) Ping(ctx context.Context) error {
	start := time.Now()
	err := c.client.Ping(ctx).Err()
	
	if err != nil {
		c.logger.WithFields(logrus.Fields{
			"error":    err,
			"duration": time.Since(start),
		}).Error("Redis ping failed")
		return err
	}

	c.logger.WithField("duration", time.Since(start)).Debug("Redis ping successful")
	return nil
}

// Close closes the Redis connection
func (c *CacheClient) Close() error {
	c.logger.Info("Closing Redis connection")
	return c.client.Close()
}

// getLogFields creates standard logging fields for cache operations
func (c *CacheClient) getLogFields(key string) *logrus.Entry {
	return c.logger.WithFields(logrus.Fields{
		"prefix": c.prefix,
		"key":    key,
	})
}