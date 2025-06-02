package cache

import (
	"context"
	"testing"
	"time"

	"github.com/himdhiman/dashboard-backend/libs/logger"
	"github.com/stretchr/testify/assert"
)

func TestMemoryCache_SetAndGet(t *testing.T) {
	logger := logger.New(nil)
	config := &CacheConfig{Timeout: 5 * time.Second}
	cache, err := NewMemoryCache(config, logger)
	if err != nil {
		t.Fatalf("Failed to create memory cache: %v", err)
	}

	// Test setting and getting a value
	ctx := context.Background()
	err = cache.Set(ctx, "key1", "value1", 0)
	assert.NoError(t, err)

	var value string
	err = cache.Get(ctx, "key1", &value)
	assert.NoError(t, err)
	assert.Equal(t, "value1", value)

	// Test getting a non-existent key
	var nonExistentValue string
	err = cache.Get(ctx, "nonExistentKey", &nonExistentValue)
	assert.Error(t, err)
}

func TestMemoryCache_Expiration(t *testing.T) {
	logger := logger.New(nil)
	config := &CacheConfig{Timeout: 1 * time.Second}
	cache, err := NewMemoryCache(config, logger)
	if err != nil {
		t.Fatalf("Failed to create memory cache: %v", err)
	}

	// Test setting a value with expiration
	ctx := context.Background()
	err = cache.Set(ctx, "key1", "value1", 500*time.Millisecond)
	assert.NoError(t, err)

	time.Sleep(600 * time.Millisecond)

	var value string
	err = cache.Get(ctx, "key1", &value)
	assert.Error(t, err) // Key should have expired
}

func TestMemoryCache_Delete(t *testing.T) {
	logger := logger.New(nil)
	config := &CacheConfig{Timeout: 5 * time.Second}
	cache, err := NewMemoryCache(config, logger)
	if err != nil {
		t.Fatalf("Failed to create memory cache: %v", err)
	}

	// Test deleting a key
	ctx := context.Background()

	err = cache.Set(ctx, "key1", "value1", 0)
	assert.NoError(t, err)

	err = cache.Delete(ctx, "key1")
	assert.NoError(t, err)

	var value string
	err = cache.Get(ctx, "key1", &value)
	assert.Error(t, err) // Key should not exist
}

func TestMemoryCache_Exists(t *testing.T) {
	logger := logger.New(nil)
	config := &CacheConfig{Timeout: 5 * time.Second}
	cache, err := NewMemoryCache(config, logger)
	if err != nil {
		t.Fatalf("Failed to create memory cache: %v", err)
	}

	// Test checking existence of a key
	ctx := context.Background()

	err = cache.Set(ctx, "key1", "value1", 0)
	assert.NoError(t, err)

	exists, err := cache.Exists(ctx, "key1")
	assert.NoError(t, err)
	assert.True(t, exists)

	exists, err = cache.Exists(ctx, "nonExistentKey")
	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestMemoryCache_IncrementDecrement(t *testing.T) {
	logger := logger.New(nil)
	config := &CacheConfig{Timeout: 5 * time.Second}
	cache, err := NewMemoryCache(config, logger)
	if err != nil {
		t.Fatalf("Failed to create memory cache: %v", err)
	}

	// Test incrementing a value
	ctx := context.Background()

	err = cache.Set(ctx, "counter", int64(10), 0)
	assert.NoError(t, err)

	newValue, err := cache.Increment(ctx, "counter")
	assert.NoError(t, err)
	assert.Equal(t, int64(11), newValue)

	// Test decrementing a value
	newValue, err = cache.Decrement(ctx, "counter")
	assert.NoError(t, err)
	assert.Equal(t, int64(10), newValue)
}

func TestMemoryCache_SetNX(t *testing.T) {
	logger := logger.New(nil)
	config := &CacheConfig{Timeout: 5 * time.Second}
	cache, err := NewMemoryCache(config, logger)
	if err != nil {
		t.Fatalf("Failed to create memory cache: %v", err)
	}

	// Test setting a value if the key does not exist
	ctx := context.Background()

	success, err := cache.SetNX(ctx, "key1", "value1", 0)
	assert.NoError(t, err)
	assert.True(t, success)

	// Test setting a value if the key already exists
	success, err = cache.SetNX(ctx, "key1", "value2", 0)
	assert.NoError(t, err)
	assert.False(t, success)

	var value string
	err = cache.Get(ctx, "key1", &value)
	assert.NoError(t, err)
	assert.Equal(t, "value1", value) // Value should not have been overwritten
}

func TestMemoryCache_GetSet(t *testing.T) {
	logger := logger.New(nil)
	config := &CacheConfig{Timeout: 5 * time.Second}
	cache, err := NewMemoryCache(config, logger)
	if err != nil {
		t.Fatalf("Failed to create memory cache: %v", err)
	}

	// Test atomically getting and setting a value
	ctx := context.Background()

	err = cache.Set(ctx, "key1", "value1", 0)
	assert.NoError(t, err)

	var oldValue string
	err = cache.GetSet(ctx, "key1", "value2", &oldValue)
	assert.NoError(t, err)
	assert.Equal(t, "value1", oldValue)

	var newValue string
	err = cache.Get(ctx, "key1", &newValue)
	assert.NoError(t, err)
	assert.Equal(t, "value2", newValue)
}

// func TestRedisCache_SetAndGet(t *testing.T) {
//     // Assuming a running Redis instance on localhost:6379
//     cache := NewRedisCache(WithRedisAddress("localhost:6379"))

//     // Test setting and getting a value
//     err := cache.Set("key1", "value1", 0)
//     assert.NoError(t, err)

//     value, err := cache.Get("key1")
//     assert.NoError(t, err)
//     assert.Equal(t, "value1", value)

//     // Test getting a non-existent key
//     _, err = cache.Get("nonExistentKey")
//     assert.Error(t, err)
// }

// func TestRedisCache_Expiration(t *testing.T) {
//     cache := NewRedisCache(WithRedisAddress("localhost:6379"))

//     // Test setting a value with expiration
//     err := cache.Set("key1", "value1", 1*time.Second)
//     assert.NoError(t, err)

//     time.Sleep(2 * time.Second)

//     _, err = cache.Get("key1")
//     assert.Error(t, err) // Key should have expired
// }

// func TestRedisCache_Delete(t *testing.T) {
//     cache := NewRedisCache(WithRedisAddress("localhost:6379"))

//     // Test deleting a key
//     err := cache.Set("key1", "value1", 0)
//     assert.NoError(t, err)

//     err = cache.Delete("key1")
//     assert.NoError(t, err)

//     _, err = cache.Get("key1")
//     assert.Error(t, err) // Key should not exist
// }

// func TestRedisCache_Exists(t *testing.T) {
//     cache := NewRedisCache(WithRedisAddress("localhost:6379"))

//     // Test checking existence of a key
//     err := cache.Set("key1", "value1", 0)
//     assert.NoError(t, err)

//     exists := cache.Exists("key1")
//     assert.True(t, exists)

//     exists = cache.Exists("nonExistentKey")
//     assert.False(t, exists)
// }
