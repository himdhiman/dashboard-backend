package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/himdhiman/dashboard-backend/libs/logger"
)

type BasicAuthStrategy struct {
	Username string
	Password string
	AuthURL  string
	Logger   logger.ILogger
}

// NewBasicAuthStrategy initializes a new instance of BasicAuthStrategy.
func NewBasicAuthStrategy(username, password, authURL string, logger logger.ILogger) *BasicAuthStrategy {
	return &BasicAuthStrategy{
		Username: username,
		Password: password,
		AuthURL:  authURL,
		Logger:   logger,
	}
}

// FetchTokens fetches new access and refresh tokens using Basic Authentication.
func (b *BasicAuthStrategy) FetchTokens(ctx context.Context) (*TokenMetadata, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, b.AuthURL, nil)
	if err != nil {
		b.Logger.Error("Error creating request for FetchTokens", "error", err)
		return nil, err
	}

	req.SetBasicAuth(b.Username, b.Password)
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		b.Logger.Error("Error making request to fetch tokens", "error", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b.Logger.Warn("Received non-200 response while fetching tokens", "status", resp.StatusCode)
		return nil, fmt.Errorf("failed to fetch tokens: %v", resp.StatusCode)
	}

	var tokenResponse TokenMetadata

	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		b.Logger.Error("Error decoding token response", "error", err)
		return nil, err
	}

	if tokenResponse.AccessToken == "" || tokenResponse.RefreshToken == "" {
		return nil, errors.New("invalid token response from server")
	}

	b.Logger.Info("Successfully fetched tokens")
	return &tokenResponse, nil
}

// RefreshTokens refreshes access and refresh tokens using the provided refresh token.
// func (b *BasicAuthStrategy) RefreshTokens(ctx context.Context, refreshToken string) (string, string, time.Time, error) {
// 	req, err := http.NewRequestWithContext(ctx, http.MethodPost, b.AuthURL+"/refresh", nil)
// 	if err != nil {
// 		b.Logger.Error("Error creating request for RefreshTokens", "error", err)
// 		return "", "", time.Time{}, err
// 	}

// 	req.SetBasicAuth(b.Username, b.Password)

// 	req.Header.Set("Authorization", "Bearer "+refreshToken)
// }
