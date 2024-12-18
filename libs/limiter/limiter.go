// internal/limiter/limiter.go
package limiter

import (
	"context"
	"errors"

	"github.com/himdhiman/dashboard-backend/libs/redis_cache"
	"github.com/himdhiman/dashboard-backend/libs/logger"
)

type RateLimiter struct {
	redisClient *redis_cache.Client
	logger      logger.Logger
}

func NewRateLimiter(redisClient *redis_cache.Client, logger logger.Logger) *RateLimiter {
	return &RateLimiter{redisClient: redisClient, logger: logger}
}

func (r *RateLimiter) Allow(ctx context.Context, serviceName, endpoint string) (bool, error) {
	// Check endpoint-specific config
	cacheKey := "rate_limit:" + serviceName + ":" + endpoint
	var endpointConfig EndpointConfig
	err := r.redisClient.Get(ctx, cacheKey, &endpointConfig)
	if err == nil {
		return r.applyRateLimit(ctx, endpointConfig)
	}

	// Fallback to default config
	defaultCacheKey := "rate_limit:" + serviceName + ":default"
	var defaultConfig EndpointConfig
	err = r.redisClient.Get(ctx, defaultCacheKey, &defaultConfig)
	if err != nil {
		return true, nil
	}

	return r.applyRateLimit(ctx, defaultConfig)
}

func (r *RateLimiter) applyRateLimit(ctx context.Context, config EndpointConfig) (bool, error) {
	switch config.Algorithm {
	case FixedWindow:
		return r.fixedWindow(ctx, config.Endpoint, config)
	case SlidingWindow:
		return r.slidingWindow(ctx, config.Endpoint, config)
	case TokenBucket:
		return r.tokenBucket(ctx, config.Endpoint, config)
	case LeakyBucket:
		return r.leakyBucket(ctx, config.Endpoint, config)
	default:
		return false, errors.New("unsupported rate limiting algorithm")
	}
}

// Add algorithm-specific implementations for fixedWindow, slidingWindow, tokenBucket, and leakyBucket.
