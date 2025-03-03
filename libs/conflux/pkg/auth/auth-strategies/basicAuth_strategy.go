package authstrategies

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/himdhiman/dashboard-backend/libs/conflux/pkg/models"
	"github.com/himdhiman/dashboard-backend/libs/logger"
)

type BasicAuthStrategy struct {
	Credentials string
	Logger   logger.ILogger
}

// NewBasicAuthStrategy initializes a new instance of BasicAuthStrategy.
func NewBasicAuthStrategy(credentials string, logger logger.ILogger) *BasicAuthStrategy {
	return &BasicAuthStrategy{
		Credentials: credentials,
		Logger:   logger,
	}
}

// FetchTokens fetches new access and refresh tokens using Basic Authentication.
func (b *BasicAuthStrategy) FetchTokens(ctx context.Context) (*models.TokenMetadata, error) {
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

	var tokenResponse models.TokenMetadata

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
