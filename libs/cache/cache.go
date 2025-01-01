package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/himdhiman/dashboard-backend/libs/logger"
)

// CacheClient implements the Cacher interface with logging
type CacheClient struct {
	client         *redis.Client
	defaultTimeout time.Duration
	prefix         string
	logger         logger.LoggerInterface
}

type CacheConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
	Timeout  time.Duration
	Prefix   string
}

// NewCacheConfig creates a new cache configuration
func NewCacheConfig(host string, port int, password string, db int, timeout time.Duration, prefix string) *CacheConfig {
	return &CacheConfig{
		Host:     host,
		Port:     port,
		Password: password,
		DB:       db,
		Timeout:  timeout,
		Prefix:   prefix,
	}
}

// NewCacheClient creates a new Redis cache client with optional configurations
func NewCacheClient(
	config *CacheConfig,
	loggerInstance logger.LoggerInterface,
	options ...CacheOption,
) *CacheClient {
	// Validate logger
	if loggerInstance == nil {
		loggerInstance = logger.New(logger.DefaultConfig())
		loggerInstance.Warn("No logger provided, using default logger")
	}

	// Create Redis client configuration
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password: config.Password,
		DB:       config.DB,
	})

	// Create cache client
	cache := &CacheClient{
		client:         rdb,
		defaultTimeout: config.Timeout * time.Hour,        // Default timeout
		prefix:         fmt.Sprintf("%s:", config.Prefix), // Default prefix
		logger:         loggerInstance,
	}

	// Apply optional configurations
	for _, opt := range options {
		opt(cache)
	}

	// Log cache client initialization
	cache.logger.WithFields(logger.Fields{
		"host":     config.Host,
		"port":     config.Port,
		"database": config.DB,
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
		c.logger.WithFields(logger.Fields{
			"error":    err,
			"duration": time.Since(start),
		}).Error("Redis ping failed")
		return err
	}

	c.logger.WithFields(logger.Fields{"duration": time.Since(start)}).Debug("Redis ping successful")
	return nil
}

// Close closes the Redis connection
func (c *CacheClient) Close() error {
	c.logger.Info("Closing Redis connection")
	return c.client.Close()
}

// getLogFields creates standard logging fields for cache operations
func (c *CacheClient) getLogFields(key string) logger.Fields {
	return logger.Fields{
		"key":    key,
		"prefix": c.prefix,
	}
}
