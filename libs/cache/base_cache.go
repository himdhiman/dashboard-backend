package cache

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/himdhiman/dashboard-backend/libs/logger"
)

// BaseCache provides common functionality for different cache implementations
type BaseCache struct {
	prefix         string
	defaultTimeout time.Duration
	logger         logger.ILogger
}

// NewBaseCache creates a new base cache instance
func NewBaseCache(prefix string, timeout time.Duration, loggerInstance logger.ILogger) *BaseCache {
	if loggerInstance == nil {
		loggerInstance = logger.New(logger.DefaultConfig("Cache Logger"))
	}

	return &BaseCache{
		prefix:         prefix + ":",
		defaultTimeout: timeout,
		logger:         loggerInstance,
	}
}

// buildKey constructs the full cache key with prefix
func (b *BaseCache) buildKey(key string) string {
	return b.prefix + key
}

func (b *BaseCache) buildKeys(keys ...string) string {
	return b.prefix + strings.Join(keys, ":")
}

// getLogFields creates standard logging fields for cache operations
func (b *BaseCache) getLogFields(key string) logger.Fields {
	return logger.Fields{
		"key":    key,
		"prefix": b.prefix,
	}
}

// serializeValue serializes a value for storage
func (b *BaseCache) serializeValue(value interface{}) ([]byte, error) {
	switch v := value.(type) {
	case string:
		return []byte(v), nil
	case []byte:
		return v, nil
	default:
		return json.Marshal(v)
	}
}

// deserializeValue deserializes a value into the result type
func (b *BaseCache) deserializeValue(data []byte, result interface{}) error {
	switch v := result.(type) {
	case *string:
		*v = string(data)
		return nil
	case *[]byte:
		*v = data
		return nil
	default:
		return json.Unmarshal(data, result)
	}
}
