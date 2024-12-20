package middleware

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/himdhiman/dashboard-backend/libs/cache"
	"github.com/himdhiman/dashboard-backend/libs/logger"
	"github.com/himdhiman/dashboard-backend/services/sentinel-service/models"
)

type AuthenticationMiddleware struct {
	Cache  *cache.CacheClient
	Logger logger.LoggerInterface
	Mutex  sync.RWMutex
}

func NewAuthenticationMiddleware(cache *cache.CacheClient, logger logger.LoggerInterface) *AuthenticationMiddleware {
	return &AuthenticationMiddleware{
		Cache:  cache,
		Logger: logger,
	}
}

func (am *AuthenticationMiddleware) Authenticate(apiName string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		exists, err := am.cacheExists(ctx, apiName+":endpoint")
		if err != nil {
			am.handleInternalError(w, "Error while fetching the config from Redis", err)
			return
		}

		if !exists {
			am.Logger.Warn("No authentication config found for endpoint:", "endpoint", r.URL.Path)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Fetch authentication type
		var authType string
		if err := am.cacheGet(ctx, apiName+":auth_type", &authType); err != nil {
			am.handleInternalError(w, "Error while fetching auth type from Redis", err)
			return
		}

		// Apply authentication
		if err := am.applyAuth(r, apiName, authType); err != nil {
			am.Logger.Error("Authentication failed:", "error", err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}


// applyAuth selects and applies the appropriate authentication method.
func (am *AuthenticationMiddleware) applyAuth(r *http.Request, apiName, authType string) error {
	switch authType {
	case models.OAUTH:
		return am.handleOAuth2(r, apiName)
	case models.BASIC_AUTH:
		return am.handleBasicAuth(r, apiName)
	default:
		return fmt.Errorf("unsupported authentication type: %s", authType)
	}
}

// handleOAuth2 applies OAuth2 authentication.
func (am *AuthenticationMiddleware) handleOAuth2(r *http.Request, apiName string) error {
	// Replace with real token-fetching logic
	token := "dummy_token_" + apiName
	r.Header.Set("Authorization", "Bearer "+token)
	return nil
}


// handleBasicAuth applies Basic Authentication using credentials from cache.
func (am *AuthenticationMiddleware) handleBasicAuth(r *http.Request, apiName string) error {
	var username, password string

	if err := am.cacheGet(context.Background(), apiName+":username", &username); err != nil {
		return fmt.Errorf("failed to fetch username: %w", err)
	}

	if err := am.cacheGet(context.Background(), apiName+":password", &password); err != nil {
		return fmt.Errorf("failed to fetch password: %w", err)
	}

	r.SetBasicAuth(username, password)
	return nil
}


// cacheExists checks if a key exists in cache.
func (am *AuthenticationMiddleware) cacheExists(ctx context.Context, key string) (bool, error) {
	am.Mutex.RLock()
	defer am.Mutex.RUnlock()
	return am.Cache.Exists(ctx, key)
}

// cacheGet retrieves a value from the cache.
func (am *AuthenticationMiddleware) cacheGet(ctx context.Context, key string, dest interface{}) error {
	am.Mutex.RLock()
	defer am.Mutex.RUnlock()

	if err := am.Cache.Get(ctx, key, dest); err != nil {
		return fmt.Errorf("cache get failed for key %s: %w", key, err)
	}
	return nil
}

// handleInternalError logs the error and sends a 500 response.
func (am *AuthenticationMiddleware) handleInternalError(w http.ResponseWriter, message string, err error) {
	am.Logger.Error(message, "error", err)
	http.Error(w, "Internal Server Error", http.StatusInternalServerError)
}
