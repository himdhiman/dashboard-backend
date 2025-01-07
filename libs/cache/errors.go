package cache

import "fmt"

type ErrorCode string

const (
	ErrCodeCacheMiss    ErrorCode = "CACHE_MISS"
	ErrCodeCacheSet     ErrorCode = "CACHE_SET"
	ErrCodeCacheDelete  ErrorCode = "CACHE_DELETE"
	ErrCodeCacheInvalid ErrorCode = "CACHE_INVALID"
	ErrCodeCacheConnect ErrorCode = "CACHE_CONNECT"
)

type CacheError struct {
	Code    ErrorCode
	Message string
	Err     error
}

func (e *CacheError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func NewCacheMissError(key string) *CacheError {
	return &CacheError{
		Code:    ErrCodeCacheMiss,
		Message: fmt.Sprintf("key not found: %s", key),
	}
}

func NewCacheSetError(key string, err error) *CacheError {
	return &CacheError{
		Code:    ErrCodeCacheSet,
		Message: fmt.Sprintf("failed to set key: %s", key),
		Err:     err,
	}
}

func NewCacheDeleteError(key string, err error) *CacheError {
	return &CacheError{
		Code:    ErrCodeCacheDelete,
		Message: fmt.Sprintf("failed to delete key: %s", key),
		Err:     err,
	}
}

func NewCacheInvalidError(msg string) *CacheError {
	return &CacheError{
		Code:    ErrCodeCacheInvalid,
		Message: msg,
	}
}

func NewCacheConnectError(err error) *CacheError {
	return &CacheError{
		Code:    ErrCodeCacheConnect,
		Message: "failed to connect to cache",
		Err:     err,
	}
}

func IsCacheMiss(err error) bool {
	if cacheErr, ok := err.(*CacheError); ok {
		return cacheErr.Code == ErrCodeCacheMiss
	}
	return false
}

func NewCacheExistsError(key string, err error) error {
	return &CacheError{
		Code:    ErrCodeCacheMiss,
		Message: fmt.Sprintf("failed to check if key exists: %s", key),
		Err:     err,
	}
}

func NewCacheDeserializeError(err error) *CacheError {
	return &CacheError{
		Code:    ErrCodeCacheMiss,
		Message: "failed to deserialize value",
		Err:     err,
	}
}
