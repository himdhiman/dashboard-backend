package cache

import (
	"context"
	"sync"
	"time"

	"github.com/himdhiman/dashboard-backend/libs/logger"
)

type cacheEntry struct {
	value      []byte
	expiration *time.Time
}

type MemoryCache struct {
	*BaseCache
	data    map[string]cacheEntry
	mu      sync.RWMutex
	cleaner *time.Ticker
}

func NewMemoryCache(config *CacheConfig, loggerInstance logger.ILogger) (Cacher, error) {
	mc := &MemoryCache{
		BaseCache: NewBaseCache(config.Prefix, config.Timeout, loggerInstance),
		data:      make(map[string]cacheEntry),
		cleaner:   time.NewTicker(time.Minute),
	}

	go mc.cleanup()
	return mc, nil
}

func (m *MemoryCache) cleanup() {
	for range m.cleaner.C {
		m.mu.Lock()
		now := time.Now()
		for k, v := range m.data {
			if v.expiration != nil && now.After(*v.expiration) {
				delete(m.data, k)
			}
		}
		m.mu.Unlock()
	}
}

func (m *MemoryCache) Set(ctx context.Context, key string, value interface{}, expiration ...time.Duration) error {
	data, err := m.serializeValue(value)
	if err != nil {
		return NewCacheInvalidError("failed to serialize value")
	}

	exp := m.defaultTimeout
	if len(expiration) > 0 {
		exp = expiration[0]
	}

	var expTime *time.Time
	if exp > 0 {
		t := time.Now().Add(exp)
		expTime = &t
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.data[m.buildKey(key)] = cacheEntry{
		value:      data,
		expiration: expTime,
	}
	return nil
}

func (m *MemoryCache) Get(ctx context.Context, key string, result interface{}) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	entry, exists := m.data[m.buildKey(key)]
	if !exists {
		return NewCacheMissError(key)
	}

	if entry.expiration != nil && time.Now().After(*entry.expiration) {
		delete(m.data, m.buildKey(key))
		return NewCacheMissError(key)
	}

	return m.deserializeValue(entry.value, result)
}

func (m *MemoryCache) GetMulti(ctx context.Context, keys []string, result interface{}) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	generatedKey := m.buildKeys(keys...)

	entry, exists := m.data[generatedKey]
	if !exists {
		return NewCacheMissError(generatedKey)
	}

	if entry.expiration != nil && time.Now().After(*entry.expiration) {
		delete(m.data, generatedKey)
		return NewCacheMissError(generatedKey)
	}

	return m.deserializeValue(entry.value, result)
}

func (m *MemoryCache) Delete(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, m.buildKey(key))
	return nil
}

func (m *MemoryCache) Exists(ctx context.Context, key string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	entry, exists := m.data[m.buildKey(key)]
	if !exists {
		return false, nil
	}

	if entry.expiration != nil && time.Now().After(*entry.expiration) {
		return false, nil
	}

	return true, nil
}

func (m *MemoryCache) Close() error {
	m.cleaner.Stop()
	m.mu.Lock()
	m.data = nil
	m.mu.Unlock()
	return nil
}

func (m *MemoryCache) Ping(ctx context.Context) error {
	return nil // Memory cache is always available
}

func (m *MemoryCache) Increment(ctx context.Context, key string) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	fullKey := m.buildKey(key)
	var value int64 = 0

	if entry, exists := m.data[fullKey]; exists {
		m.deserializeValue(entry.value, &value)
	}

	value++
	data, _ := m.serializeValue(value)
	m.data[fullKey] = cacheEntry{value: data}
	return value, nil
}

func (m *MemoryCache) Decrement(ctx context.Context, key string) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	fullKey := m.buildKey(key)
	var value int64 = 0

	if entry, exists := m.data[fullKey]; exists {
		m.deserializeValue(entry.value, &value)
	}

	value--
	data, _ := m.serializeValue(value)
	m.data[fullKey] = cacheEntry{value: data}
	return value, nil
}

func (m *MemoryCache) SetNX(ctx context.Context, key string, value interface{}, expiration ...time.Duration) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	fullKey := m.buildKey(key)
	if _, exists := m.data[fullKey]; exists {
		return false, nil
	}

	data, err := m.serializeValue(value)
	if err != nil {
		return false, NewCacheInvalidError("failed to serialize value")
	}

	exp := m.defaultTimeout
	if len(expiration) > 0 {
		exp = expiration[0]
	}

	var expTime *time.Time
	if exp > 0 {
		t := time.Now().Add(exp)
		expTime = &t
	}

	m.data[fullKey] = cacheEntry{
		value:      data,
		expiration: expTime,
	}
	return true, nil
}

func (m *MemoryCache) GetSet(ctx context.Context, key string, value interface{}, result interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	fullKey := m.buildKey(key)
	if oldEntry, exists := m.data[fullKey]; exists {
		m.deserializeValue(oldEntry.value, result)
	}

	data, err := m.serializeValue(value)
	if err != nil {
		return NewCacheInvalidError("failed to serialize value")
	}

	m.data[fullKey] = cacheEntry{value: data}
	return nil
}
