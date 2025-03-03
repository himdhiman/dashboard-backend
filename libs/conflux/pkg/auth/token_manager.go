package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/himdhiman/dashboard-backend/libs/cache"
	"github.com/himdhiman/dashboard-backend/libs/conflux/pkg/models"
	interfaces "github.com/himdhiman/dashboard-backend/libs/conflux/pkg/interface"
	"github.com/himdhiman/dashboard-backend/libs/crypto"
	"github.com/himdhiman/dashboard-backend/libs/logger"
)

type TokenManager struct {
	Mutex        sync.RWMutex
	Cache        cache.Cacher
	Crypto       *crypto.Crypto
	Logger       logger.ILogger
	ApiName      string
	AuthStrategy interfaces.AuthenticationStrategy
}

func NewTokenManager(cache cache.Cacher, logger logger.ILogger, crypto *crypto.Crypto, apiName string, strategy interfaces.AuthenticationStrategy) *TokenManager {
	logger.Info("Initializing new TokenManager for API: ", apiName)
	return &TokenManager{
		Cache:        cache,
		Logger:       logger,
		Crypto:       crypto,
		ApiName:      apiName,
		AuthStrategy: strategy,
	}
}

func (tm *TokenManager) AuthenticateRequest(ctx context.Context, r *http.Request) error {
	tm.Logger.Info("Starting authentication request for API: ", tm.ApiName)

	// Get tokens from the cache
	tokenData, err := tm.GetTokenFromCache(ctx, tm.ApiName)
	if err != nil {
		tm.Logger.Warn("Token data not found in cache; fetching new tokens")
		return tm.fetchAndAuthenticate(ctx, r)
	}

	// Check if the access token is available in the cache and not expired
	tokenExpiryTime := tokenData.CreatedAt.Add(time.Second * time.Duration(tokenData.ExpiresIn))
    if time.Now().Before(tokenExpiryTime) {
        tm.Logger.Info("Using valid access token from cache")
        r.Header.Set("Authorization", "Bearer " + tokenData.AccessToken)
        tm.Logger.Debug("Successfully set Authorization header with access token")
        return nil
    }

	// Access token is not available, try using the refresh token
	if tokenData.RefreshToken != "" {
		tm.Logger.Warn("Access token not found; attempting to refresh")
		err := tm.refreshAndAuthenticate(ctx, r)
		if err == nil {
			tm.Logger.Info("Successfully refreshed and authenticated token")
			return nil
		}
		tm.Logger.Error("Failed to refresh token: ", err)
	}

	// Refresh token also expired or unavailable, fetch new tokens
	tm.Logger.Warn("Refresh token expired or unavailable; fetching new tokens")
	return tm.fetchAndAuthenticate(ctx, r)
}

func (tm *TokenManager) fetchAndAuthenticate(ctx context.Context, r *http.Request) error {
	tm.Logger.Info("Fetching new tokens for API: ", tm.ApiName)
	tokens, err := tm.AuthStrategy.FetchTokens(ctx, tm.ApiName)
	if err != nil {
		tm.Logger.Error("Failed to fetch new tokens: ", err)
		return err
	}

	// Add access token to the request
	r.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
	tm.Logger.Debug("Successfully set Authorization header with new access token")
	return nil
}

// GetTokensFromCache retrieves token data from the cache.
func (tm *TokenManager) GetTokenFromCache(ctx context.Context, apiName string) (*models.TokenResponse, error) {
	tm.Logger.Debug("Attempting to retrieve tokens from cache for API: ", apiName)

	var encryptedToken string
	if err := tm.Cache.GetMulti(ctx, []string{apiName, "Token"}, &encryptedToken); err != nil {
		tm.Logger.Debug("Token not found in cache: ", err)
		return nil, err
	}

	decryptedToken, err := tm.Crypto.Decrypt(encryptedToken)
	if err != nil {
		tm.Logger.Error("Failed to decrypt token: ", err)
		return nil, err
	}

	var token models.TokenResponse
	if err := json.Unmarshal([]byte(decryptedToken), &token); err != nil {
		tm.Logger.Error("Failed to unmarshal token: ", err)
		return nil, err
	}

	tm.Logger.Debug("Successfully retrieved and decrypted tokens from cache")
	return &token, nil
}

func (tm *TokenManager) refreshAndAuthenticate(ctx context.Context, r *http.Request) error {
	tm.Logger.Info("Attempting to refresh tokens for API: ", tm.ApiName)
	tokens, err := tm.AuthStrategy.RefreshTokens(ctx, tm.ApiName)
	if err != nil {
		tm.Logger.Error("Failed to refresh tokens: ", err)
		return err
	}

	r.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
	tm.Logger.Debug("Successfully set Authorization header with refreshed access token")
	return nil
}
