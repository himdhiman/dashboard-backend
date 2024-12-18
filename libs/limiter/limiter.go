package limiter

import (
	"context"
	"errors"
	"time"

	"github.com/himdhiman/dashboard-backend/libs/logger"
	"github.com/himdhiman/dashboard-backend/libs/redis_cache"
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

func (r *RateLimiter) fixedWindow(ctx context.Context, key string, config EndpointConfig) (bool, error) {
	windowKey := "fixed_window:" + key
	currentCount, err := r.redisClient.Increment(ctx, windowKey, 1, config.TimeWindow)
	if err != nil {
		r.logger.Error("Fixed Window Error:", err)
		return false, err
	}

	if currentCount > config.MaxRequests {
		r.logger.Warn("Fixed Window Limit Exceeded for:", key)
		return false, nil
	}

	return true, nil
}

func (r *RateLimiter) slidingWindow(ctx context.Context, key string, config EndpointConfig) (bool, error) {
	windowKey := "sliding_window:" + key
	currentTime := time.Now().Unix()
	startTime := currentTime - int64(config.TimeWindow.Seconds())

	// Add current timestamp to the sorted set
	err := r.redisClient.ZAdd(ctx, windowKey, float64(currentTime), currentTime)
	if err != nil {
		r.logger.Error("Sliding Window Error:", err)
		return false, err
	}

	// Remove timestamps outside the window
	err = r.redisClient.ZRemRangeByScore(ctx, windowKey, 0, float64(startTime))
	if err != nil {
		r.logger.Error("Sliding Window Cleanup Error:", err)
		return false, err
	}

	// Get the current count
	count, err := r.redisClient.ZCount(ctx, windowKey, float64(startTime), float64(currentTime))
	if err != nil {
		r.logger.Error("Sliding Window Count Error:", err)
		return false, err
	}

	if int(count) > config.MaxRequests {
		r.logger.Warn("Sliding Window Limit Exceeded for:", key)
		return false, nil
	}

	// Set expiration for the key
	r.redisClient.Expire(ctx, windowKey, config.TimeWindow)

	return true, nil
}

func (r *RateLimiter) tokenBucket(ctx context.Context, key string, config EndpointConfig) (bool, error) {
	bucketKey := "token_bucket:" + key
	lastRefillKey := bucketKey + ":last_refill"
	maxTokens := config.MaxRequests
	refillRate := float64(maxTokens) / config.TimeWindow.Seconds()

	// Get the current token count and last refill time
	tokenCount, err := r.redisClient.GetInt(ctx, bucketKey)
	if err != nil {
		tokenCount = maxTokens // Initialize the bucket if not present
	}

	lastRefill, err := r.redisClient.GetInt(ctx, lastRefillKey)
	if err != nil {
		lastRefill = int(time.Now().Unix())
	}

	// Calculate the number of tokens to refill
	currentTime := time.Now().Unix()
	elapsedTime := currentTime - int64(lastRefill)
	refilledTokens := int(float64(elapsedTime) * refillRate)
	tokenCount = min(maxTokens, tokenCount+refilledTokens)

	// Save the updated token count and last refill time
	err = r.redisClient.SetInt(ctx, bucketKey, tokenCount, config.TimeWindow)
	if err != nil {
		r.logger.Error("Token Bucket Save Error:", err)
		return false, err
	}

	err = r.redisClient.SetInt(ctx, lastRefillKey, int(currentTime), config.TimeWindow)
	if err != nil {
		r.logger.Error("Token Bucket Last Refill Save Error:", err)
		return false, err
	}

	// Check if the request can be processed
	if tokenCount <= 0 {
		r.logger.Warn("Token Bucket Limit Exceeded for:", key)
		return false, nil
	}

	// Deduct a token for the request
	err = r.redisClient.Decrement(ctx, bucketKey)
	if err != nil {
		r.logger.Error("Token Bucket Deduction Error:", err)
		return false, err
	}

	return true, nil
}

func (r *RateLimiter) leakyBucket(ctx context.Context, key string, config EndpointConfig) (bool, error) {
	bucketKey := "leaky_bucket:" + key
	lastLeakKey := bucketKey + ":last_leak"
	maxCapacity := config.MaxRequests
	leakRate := float64(maxCapacity) / config.TimeWindow.Seconds()

	// Get the current bucket size and last leak time
	bucketSize, err := r.redisClient.GetInt(ctx, bucketKey)
	if err != nil {
		bucketSize = 0 // Initialize the bucket if not present
	}

	lastLeak, err := r.redisClient.GetInt(ctx, lastLeakKey)
	if err != nil {
		lastLeak = int(time.Now().Unix())
	}

	// Calculate the number of leaked tokens
	currentTime := time.Now().Unix()
	elapsedTime := currentTime - int64(lastLeak)
	leakedTokens := int(float64(elapsedTime) * leakRate)
	bucketSize = max(0, bucketSize-leakedTokens)

	// Save the updated bucket size and last leak time
	err = r.redisClient.SetInt(ctx, bucketKey, bucketSize, config.TimeWindow)
	if err != nil {
		r.logger.Error("Leaky Bucket Save Error:", err)
		return false, err
	}

	err = r.redisClient.SetInt(ctx, lastLeakKey, int(currentTime), config.TimeWindow)
	if err != nil {
		r.logger.Error("Leaky Bucket Last Leak Save Error:", err)
		return false, err
	}

	// Check if the request can be processed
	if bucketSize >= maxCapacity {
		r.logger.Warn("Leaky Bucket Limit Exceeded for:", key)
		return false, nil
	}

	// Add a token for the new request
	err = r.redisClient.Increment(ctx, bucketKey, 1, config.TimeWindow)
	if err != nil {
		r.logger.Error("Leaky Bucket Addition Error:", err)
		return false, err
	}

	return true, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Add algorithm-specific implementations for fixedWindow, slidingWindow, tokenBucket, and leakyBucket.
