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
	var endpoint, path, username, clientId, clientSecret string
	if err := a.cacheGet(ctx, apiName+":endpoint", &endpoint); err != nil {
		return nil, fmt.Errorf("failed to fetch endpoint for api %s: %w", apiName, err)
	}

	if err := a.cacheGet(ctx, apiName+":path", &path); err != nil {
		return nil, fmt.Errorf("failed to fetch path for api %s: %w", apiName, err)
	}
	if err := a.cacheGet(ctx, apiName+":username", &username); err != nil {
		return nil, fmt.Errorf("failed to fetch username for api %s: %w", apiName, err)
	}
	if err := a.cacheGet(ctx, apiName+":client_id", &clientId); err != nil {
		return nil, fmt.Errorf("failed to fetch client id for api %s: %w", apiName, err)
	}
	if err := a.cacheGet(ctx, apiName+":client_secret", &clientSecret); err != nil {
		return nil, fmt.Errorf("failed to fetch client secret for api %s: %w", apiName, err)
	}

	decryptedClientId, err := a.Crypto.Decrypt(clientId)
	if err != nil {
		a.Logger.Error("Error decrypting client id", "error", err)
		return nil, err
	}

	decryptedClientSecret, err := a.Crypto.Decrypt(clientSecret)
	if err != nil {
		a.Logger.Error("Error decrypting client secret", "error", err)
		return nil, err
	}

	decryptedUsername, err := a.Crypto.Decrypt(username)
	if err != nil {
		a.Logger.Error("Error decrypting username", "error", err)
		return nil, err
	}

	authURL, err := url.Parse(endpoint + path)
	if err != nil {
		a.Logger.Error("Error parsing URL", "error", err)
		return nil, err
	}

	// Add query parameters
	params := url.Values{}
	params.Add("grant_type", "password")
	params.Add("client_id", string(decryptedClientId))
	params.Add("username", string(decryptedUsername))
	params.Add("password", string(decryptedClientSecret))

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

	a.Logger.Info("Successfully fetched tokens")
	return &tokenResponse, nil

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
