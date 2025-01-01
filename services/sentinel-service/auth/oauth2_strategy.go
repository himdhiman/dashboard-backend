package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/himdhiman/dashboard-backend/libs/logger"
)

type OAuth2Strategy struct {
	ClientID     string
	ClientSecret string
	AuthURL      string
	TokenURL     string
	Logger       logger.ILogger
}

// NewOAuth2Strategy initializes a new instance of OAuth2Strategy.
func NewOAuth2Strategy(clientID, clientSecret, authURL, tokenURL string, logger logger.ILogger) *OAuth2Strategy {
	return &OAuth2Strategy{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		AuthURL:      authURL,
		TokenURL:     tokenURL,
		Logger:       logger,
	}
}

// FetchTokens fetches new access and refresh tokens using client credentials.
func (o *OAuth2Strategy) FetchTokens(ctx context.Context) (*TokenMetadata, error) {
	payload := map[string]string{
		"grant_type":    "client_credentials",
		"client_id":     o.ClientID,
		"client_secret": o.ClientSecret,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		o.Logger.Error("Error encoding payload for token request", "error", err)
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, o.TokenURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		o.Logger.Error("Error creating request for FetchTokens", "error", err)
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		o.Logger.Error("Error making request to fetch tokens", "error", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		o.Logger.Warn("Received non-200 response while fetching tokens", "status", resp.StatusCode)
		return nil, fmt.Errorf("failed to fetch tokens: %v", resp.StatusCode)
	}

	var tokenResponse TokenMetadata

	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		o.Logger.Error("Error decoding token response", "error", err)
		return nil, err
	}

	if tokenResponse.AccessToken == "" || tokenResponse.RefreshToken == "" {
		return nil, errors.New("invalid token response from server")
	}

	// expiryTime := time.Now().Add(time.Duration(tokenResponse.ExpiresIn) * time.Second)
	o.Logger.Info("Successfully fetched tokens")
	return &tokenResponse, nil
}

// RefreshTokens refreshes access and refresh tokens using the provided refresh token.
func (o *OAuth2Strategy) RefreshTokens(ctx context.Context, refreshToken string) (*TokenMetadata, error) {
	payload := map[string]string{
		"grant_type":    "refresh_token",
		"refresh_token": refreshToken,
		"client_id":     o.ClientID,
		"client_secret": o.ClientSecret,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		o.Logger.Error("Error encoding payload for refresh token request", "error", err)
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, o.TokenURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		o.Logger.Error("Error creating request for RefreshTokens", "error", err)
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		o.Logger.Error("Error making request to refresh tokens", "error", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		o.Logger.Warn("Received non-200 response while refreshing tokens", "status", resp.StatusCode)
		return nil, fmt.Errorf("failed to refresh tokens: %v", resp.StatusCode)
	}

	var tokenResponse TokenMetadata

	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		o.Logger.Error("Error decoding refresh token response", "error", err)
		return nil, err
	}

	if tokenResponse.AccessToken == "" || tokenResponse.RefreshToken == "" {
		return nil, errors.New("invalid refresh token response from server")
	}

	// expiryTime := time.Now().Add(time.Duration(tokenResponse.ExpiresIn) * time.Second)
	o.Logger.Info("Successfully refreshed tokens")
	return &tokenResponse, nil
}
