package auth

import (
	"context"
	"sync"

	"github.com/himdhiman/dashboard-backend/libs/cache"
	"github.com/himdhiman/dashboard-backend/libs/logger"
)

type TokenMetadata struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

type TokenManager struct {
	mu       sync.RWMutex
	cache    *cache.CacheClient
	logger   logger.LoggerInterface
	apiName  string
	strategy AuthenticationStrategy
}

func NewTokenManager(cache *cache.CacheClient, logger logger.LoggerInterface, apiName string, strategy AuthenticationStrategy) *TokenManager {
	return &TokenManager{
		cache:    cache,
		logger:   logger,
		apiName:  apiName,
		strategy: strategy,
	}
}

func (tm *TokenManager) GetAccessToken(ctx context.Context) (string, error) {
	cacheKey := tm.apiName + ":auth"
}
