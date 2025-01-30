package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/himdhiman/dashboard-backend/libs/logger"
)

type CacheType int

const (
	RedisCache CacheType = iota
	InMemoryCache
)

// CacheClient implements the Cacher interface with logging
type CacheClient struct {
	*BaseCache
	client *redis.Client
}

type CacheConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
	Timeout  time.Duration
	Prefix   string
}

// NewCacheClient creates a new Redis cache client with optional configurations
func NewCacheClient(
	config *CacheConfig,
	loggerInstance logger.ILogger,
	options ...CacheOption,
) (Cacher, error) {
	// Validate logger
	if loggerInstance == nil {
		loggerInstance = logger.New(logger.DefaultConfig("Cache Logger"))
		loggerInstance.Warn("No logger provided, using default logger")
	}

	if config == nil {
		return nil, NewCacheInvalidError("cache config is required")
	}

	// Create Redis client configuration
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password: config.Password,
		DB:       config.DB,
	})

	// Test connection
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, NewCacheConnectError(err)
	}

	// Create cache client
	cache := &CacheClient{
		BaseCache: NewBaseCache(config.Prefix, config.Timeout, loggerInstance),
		client:    rdb,
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

	return cache, nil
}

func NewCache(cacheType CacheType, config *CacheConfig, loggerInstance logger.ILogger) (Cacher, error) {
    switch cacheType {
    case RedisCache:
        return NewCacheClient(config, loggerInstance)
    case InMemoryCache:
        return NewMemoryCache(config, loggerInstance)
    default:
        return nil, NewCacheInvalidError("invalid cache type")
    }
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
