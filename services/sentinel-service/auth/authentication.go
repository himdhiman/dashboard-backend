package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/himdhiman/dashboard-backend/libs/cache"
	"github.com/himdhiman/dashboard-backend/libs/crypto"
	"github.com/himdhiman/dashboard-backend/libs/logger"

	"github.com/himdhiman/dashboard-backend/services/sentinel-service/models"
)

type Authentication struct {
	Mutex  sync.RWMutex
	Cache  *cache.CacheClient
	Crypto *crypto.Crypto
	Logger logger.LoggerInterface
}

func NewAuthentication(cache *cache.CacheClient, logger logger.LoggerInterface, crypto *crypto.Crypto) *Authentication {
	return &Authentication{
		Cache:  cache,
		Crypto: crypto,
		Logger: logger,
	}
}

func (a *Authentication) FetchTokens(ctx context.Context, apiName string) (*models.TokenResponse, error) {
	// Fetch authentication type
	var endpoint, path string
	if err := a.cacheGet(ctx, apiName+":endpoint", &endpoint); err != nil {
		return nil, fmt.Errorf("failed to fetch endpoint for api %s: %w", apiName, err)
	}

	if err := a.cacheGet(ctx, apiName+":path", &path); err != nil {
		return nil, fmt.Errorf("failed to fetch path for api %s: %w", apiName, err)
	}
	clientId, err := a.getDecryptedValue(ctx, apiName+":client_id")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch client id for api %s: %w", apiName, err)
	}

	clientSecret, err := a.getDecryptedValue(ctx, apiName+":client_secret")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch client secret for api %s: %w", apiName, err)
	}

	username, err := a.getDecryptedValue(ctx, apiName+":username")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch username for api %s: %w", apiName, err)
	}

	authURL, err := url.Parse(endpoint + path)
	if err != nil {
		a.Logger.Error("Error parsing URL", "error", err)
		return nil, err
	}

	// Add query parameters
	params := url.Values{}
	params.Add("grant_type", "password")
	params.Add("client_id", clientId)
	params.Add("username", username)
	params.Add("password", clientSecret)

	authURL.RawQuery = params.Encode()

	// Create the request with the constructed URL
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, authURL.String(), nil)
	if err != nil {
		a.Logger.Error("Error creating request for RefreshTokens", "error", err)
		return nil, err
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		a.Logger.Error("Error making request to fetch tokens", "error", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		a.Logger.Warn("Received non-200 response while fetching tokens", "status", resp.StatusCode)
		return nil, fmt.Errorf("failed to fetch tokens: %v", resp.StatusCode)
	}

	var tokenResponse models.TokenResponse

	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		a.Logger.Error("Error decoding token response", "error", err)
		return nil, err
	}

	if tokenResponse.AccessToken == "" || tokenResponse.RefreshToken == "" {
		return nil, errors.New("invalid token response from server")
	}

	a.setCacheValue(ctx, apiName+":access_token", tokenResponse.AccessToken, time.Duration(tokenResponse.ExpiresIn)*time.Second)
	a.setCacheValue(ctx, apiName+":refresh_token", tokenResponse.RefreshToken)

	a.Logger.Info("Successfully fetched tokens and stored in cache")
	return &tokenResponse, nil

}

// RefreshTokens refreshes access and refresh tokens using the provided refresh token.
func (a *Authentication) RefreshTokens(ctx context.Context, apiName string) (*models.TokenResponse, error) {
	var endpoint, path string

	if err := a.cacheGet(ctx, apiName+":endpoint", &endpoint); err != nil {
		return nil, fmt.Errorf("failed to fetch endpoint for api %s: %w", apiName, err)
	}

	if err := a.cacheGet(ctx, apiName+":path", &path); err != nil {
		return nil, fmt.Errorf("failed to fetch path for api %s: %w", apiName, err)
	}

	clientId, err := a.getDecryptedValue(ctx, apiName+":client_id")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch client id for api %s: %w", apiName, err)
	}

	refreshToken, err := a.getDecryptedValue(ctx, apiName+":refresh_token")
	if err != nil || refreshToken == "" {
		a.Logger.Warn("Refresh token not found or invalid, refetching tokens")
		return a.FetchTokens(ctx, apiName)
	}

	authURL, err := url.Parse(endpoint + path)
	if err != nil {
		a.Logger.Error("Error parsing URL", "error", err)
		return nil, err
	}

	// Add query parameters
	params := url.Values{}
	params.Add("grant_type", "refresh_token")
	params.Add("client_id", clientId)
	params.Add("refresh_token", refreshToken)

	authURL.RawQuery = params.Encode()

	// Create the request with the constructed URL
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, authURL.String(), nil)
	if err != nil {
		a.Logger.Error("Error creating request for RefreshTokens", "error", err)
		return nil, err
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		a.Logger.Error("Error making request to fetch tokens", "error", err)
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// If the refresh token is expired, refetch the tokens
		if resp.StatusCode == http.StatusUnauthorized {
			a.Logger.Warn("Refresh token expired, refetching tokens")
			a.Cache.Delete(ctx, apiName+":refresh_token")
			return a.FetchTokens(ctx, apiName)
		}
		a.Logger.Warn("Received non-200 response while refreshing tokens", "status", resp.StatusCode)
		return nil, fmt.Errorf("failed to refresh tokens: %v", resp.StatusCode)
	}

	var tokenResponse models.TokenResponse

	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		a.Logger.Error("Error decoding token response", "error", err)
		return nil, err
	}

	if tokenResponse.AccessToken == "" || tokenResponse.RefreshToken == "" {
		return nil, errors.New("invalid token response from server")
	}

	a.setCacheValue(ctx, apiName+":access_token", tokenResponse.AccessToken, time.Duration(tokenResponse.ExpiresIn)*time.Second)
	a.setCacheValue(ctx, apiName+":refresh_token", tokenResponse.RefreshToken)

	a.Logger.Info("Successfully fetched tokens and stored in cache")
	return &tokenResponse, nil
}

// getDecryptedValue retrieves and decrypts a value from the cache.
func (a *Authentication) getDecryptedValue(ctx context.Context, key string) (string, error) {
	var encryptedValue string
	if err := a.cacheGet(ctx, key, &encryptedValue); err != nil {
		return "", err
	}

	decryptedValue, err := a.Crypto.Decrypt(encryptedValue)
	if err != nil {
		a.Logger.Error("Error decrypting value", "key", key, "error", err)
		return "", err
	}

	return string(decryptedValue), nil
}

// setCacheValue encrypts and sets the value in the cache with an optional expiration.
func (a *Authentication) setCacheValue(ctx context.Context, key, value string, expiration ...time.Duration) {
	a.Mutex.Lock()
	defer a.Mutex.Unlock()

	encryptedValue, err := a.Crypto.Encrypt(value)
	if err != nil {
		a.Logger.Error("Error encrypting value", "key", key, "error", err)
		return
	}

	a.Cache.Set(ctx, key, encryptedValue, expiration...)
}

// cacheGet retrieves a value from the cache.
func (a *Authentication) cacheGet(ctx context.Context, key string, dest interface{}) error {
	a.Mutex.RLock()
	defer a.Mutex.RUnlock()

	if err := a.Cache.Get(ctx, key, dest); err != nil {
		return fmt.Errorf("cache get failed for key %s: %w", key, err)
	}
	return nil
}
