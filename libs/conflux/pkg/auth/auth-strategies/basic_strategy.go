package authstrategies

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/himdhiman/dashboard-backend/libs/cache"
	"github.com/himdhiman/dashboard-backend/libs/conflux/pkg/models"
	"github.com/himdhiman/dashboard-backend/libs/crypto"
	"github.com/himdhiman/dashboard-backend/libs/logger"
)

type BasicAuthStrategy struct {
	BaseStrategy
}

// NewBasicAuthStrategy initializes a new instance of BasicAuthStrategy.
func NewBasicAuthStrategy(apiCode string, credentials models.Credentials, authURL string, logger logger.ILogger, cache cache.Cacher, crypto *crypto.Crypto) *BasicAuthStrategy {
	return &BasicAuthStrategy{
		BaseStrategy: BaseStrategy{
			ApiCode:     apiCode,
			AuthURL:     authURL,
			Credentials: credentials,
			Logger:      logger,
			Cache:       cache,
			Crypto:      crypto,
		},
	}
}

// FetchTokens fetches new access and refresh tokens using client credentials.
func (a *BasicAuthStrategy) FetchTokens(ctx context.Context) (*models.TokenResponse, error) {

	clientId, err := a.Crypto.Decrypt(a.Credentials.ClientID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch client id for api %s: %w", a.ApiCode, err)
	}

	clientSecret, err := a.Crypto.Decrypt(a.Credentials.ClientSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch client secret for api %s: %w", a.ApiCode, err)
	}

	username, err := a.Crypto.Decrypt(a.Credentials.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch username for api %s: %w", a.ApiCode, err)
	}

	authURL, err := url.Parse(a.AuthURL)
	if err != nil {
		a.Logger.Error("Error parsing URL", "error", err)
		return nil, err
	}

	// Add query parameters
	params := url.Values{}
	params.Add("grant_type", "password")
	params.Add("client_id", string(clientId))
	params.Add("username", string(username))
	params.Add("password", string(clientSecret))

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

	tokenResponse.CreatedAt = time.Now()

	a.Logger.Info("Successfully fetched tokens and stored in cache")
	return &tokenResponse, nil

}

// RefreshTokens refreshes access and refresh tokens using the provided refresh token.
func (a *BasicAuthStrategy) RefreshTokens(ctx context.Context) (*models.TokenResponse, error) {

	// get the token from the cache
	var token models.TokenResponse
	err := a.Cache.Get(ctx, a.ApiCode+":Token", &token)
	if err != nil {
		a.Logger.Error("Error fetching token from cache", "error", err)
		return nil, err
	}

	if token.RefreshToken == "" {
		a.Logger.Warn("Refresh token not found in cache, refetching tokens")
		return a.FetchTokens(ctx)
	}

	authURL, err := url.Parse(a.AuthURL)
	if err != nil {
		a.Logger.Error("Error parsing URL", "error", err)
		return nil, err
	}

	clientId, err := a.Crypto.Decrypt(a.Credentials.ClientID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch client id for api %s: %w", a.ApiCode, err)
	}

	// Add query parameters
	params := url.Values{}
	params.Add("grant_type", "refresh_token")
	params.Add("client_id", string(clientId))
	params.Add("refresh_token", token.RefreshToken)

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
			return a.FetchTokens(ctx)
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

	tokenResponse.CreatedAt = time.Now()

	a.Logger.Info("Successfully fetched tokens and stored in cache")
	return &tokenResponse, nil
}
