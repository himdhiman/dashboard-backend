package auth

import (
	"context"
	"net/http"
	"sync"

	"github.com/himdhiman/dashboard-backend/libs/cache"
	"github.com/himdhiman/dashboard-backend/libs/crypto"
	"github.com/himdhiman/dashboard-backend/libs/logger"
	"github.com/himdhiman/dashboard-backend/services/sentinel-service/models"
)

type TokenMetadata struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

type TokenManager struct {
	Mutex        sync.RWMutex
	Cache        cache.Cacher
	Crypto       *crypto.Crypto
	Logger       logger.ILogger
	ApiName      string
	AuthStrategy AuthenticationStrategy
}

func NewTokenManager(cache cache.Cacher, logger logger.ILogger, crypto *crypto.Crypto, apiName string, strategy AuthenticationStrategy) *TokenManager {
	return &TokenManager{
		Cache:        cache,
		Logger:       logger,
		Crypto:       crypto,
		ApiName:      apiName,
		AuthStrategy: strategy,
	}
}

func (tm *TokenManager) AuthenticateRequest(ctx context.Context, r *http.Request) error {
	// Get tokens from the cache
	tokenData, err := tm.GetTokenFromCache(ctx, tm.ApiName)
	if err != nil {
		tm.Logger.Warn("Token data not found in cache; fetching new tokens")
		return tm.fetchAndAuthenticate(ctx, r)
	}

	// Check if the access token is available in the cache
	if tokenData.AccessToken != "" {
		tm.Logger.Info("Using valid access token from cache")
		r.Header.Set("Authorization", "Bearer "+tokenData.AccessToken)
		return nil
	}

	// Access token is not available, try using the refresh token
	if tokenData.RefreshToken != "" {
		tm.Logger.Warn("Access token not found; attempting to refresh")
		err := tm.refreshAndAuthenticate(ctx, r)
		if err == nil {
			return nil
		}
		tm.Logger.Error("Failed to refresh token: ", err)
	}

	// Refresh token also expired or unavailable, fetch new tokens
	tm.Logger.Warn("Refresh token expired or unavailable; fetching new tokens")
	return tm.fetchAndAuthenticate(ctx, r)
}

func (tm *TokenManager) fetchAndAuthenticate(ctx context.Context, r *http.Request) error {
	tokens, err := tm.AuthStrategy.FetchTokens(ctx, tm.ApiName)
	if err != nil {
		tm.Logger.Error("Failed to fetch new tokens: ", err)
		return err
	}

	// Add access token to the request
	r.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
	return nil
}

func (tm *TokenManager) refreshAndAuthenticate(ctx context.Context, r *http.Request) error {
	tokens, err := tm.AuthStrategy.RefreshTokens(ctx, tm.ApiName)
	if err != nil {
		tm.Logger.Error("Failed to refresh tokens: ", err)
		return err
	}

	r.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
	return nil
}

// GetTokensFromCache retrieves token data from the cache.
func (tm *TokenManager) GetTokenFromCache(ctx context.Context, apiName string) (*models.TokenResponse, error) {
	var accessToken, refreshToken string
	if err := tm.Cache.Get(ctx, apiName+":access_token", &accessToken); err != nil {
		return nil, err
	}
	if err := tm.Cache.Get(ctx, apiName+":refresh_token", &refreshToken); err != nil {
		return nil, err
	}

	decryptedAccessToken, err := tm.Crypto.Decrypt(accessToken)
	if err != nil {
		return nil, err
	}

	decryptedRefreshToken, err := tm.Crypto.Decrypt(refreshToken)
	if err != nil {
		return nil, err
	}

	return &models.TokenResponse{
		AccessToken:  string(decryptedAccessToken),
		RefreshToken: string(decryptedRefreshToken),
	}, nil
}
