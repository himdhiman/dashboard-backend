package cache

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-redis/redis/v8"
)

// Get retrieves a value from the cache and deserialize it into the result type
func (c *CacheClient) Get(
	ctx context.Context,
	key string,
	result interface{},
) *CacheError {
	fullKey := c.buildKey(key)

	// Retrieve the raw data
	data, err := c.client.Get(ctx, fullKey).Bytes()
	if err != nil {
		if err == redis.Nil {
			return NewCacheMissError(key)
		}
		return NewCacheConnectError(err)
	}

	// Deserialize based on result type
	switch v := result.(type) {
	case *string:
		*v = string(data)
		return nil
	case *[]byte:
		*v = data
		return nil
	default:
		if err := json.Unmarshal(data, result); err != nil {
			return NewCacheDeserializeError(err)
		}
		return nil
	}
}

// GetSet atomically sets a new value and returns the old value
func (c *CacheClient) GetSet(
	ctx context.Context,
	key string,
	value interface{},
	result interface{},
) error {
	// Serialize the new value
	var data []byte
	var err error
	switch v := value.(type) {
	case string:
		data = []byte(v)
	case []byte:
		data = v
	default:
		data, err = json.Marshal(v)
		if err != nil {
			return fmt.Errorf("failed to serialize value: %v", err)
		}
	}

	fullKey := c.buildKey(key)

	// Get and set atomically
	oldData, err := c.client.GetSet(ctx, fullKey, data).Bytes()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("key not found: %s", key)
		}
		return err
	}

	// Deserialize the old value
	switch v := result.(type) {
	case *string:
		*v = string(oldData)
		return nil
	case *[]byte:
		*v = oldData
		return nil
	default:
		return json.Unmarshal(oldData, result)
	}
}
