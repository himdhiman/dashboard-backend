package client

import (
	"context"
	"io/ioutil"
	"net/http"

	"github.com/himdhiman/dashboard-backend/libs/conflux/pkg/models"
	interfaces "github.com/himdhiman/dashboard-backend/libs/conflux/pkg/interface"
	"github.com/himdhiman/dashboard-backend/libs/logger"
	"github.com/himdhiman/dashboard-backend/libs/cache"
)

type ConfluxAPIClient struct {
	models.APIConfig
	auth       interfaces.AuthenticationStrategy
	httpClient *http.Client
	logger     logger.ILogger
	cache      cache.Cacher
}

// NewConfluxAPIClient creates a new instance of ConfluxAPIClient
func NewConfluxAPIClient(config models.APIConfig, auth interfaces.AuthenticationStrategy, httpClient *http.Client, logger logger.ILogger, cache cache.Cacher) *ConfluxAPIClient {
	return &ConfluxAPIClient{
		APIConfig:  config,
		auth:       auth,
		httpClient: httpClient,
		logger:     logger,
		cache:      cache,
	}
}

// GetBaseURL returns the configured BaseURL.
func (c *ConfluxAPIClient) GetBaseURL() string {
	return c.APIConfig.BaseURL
}

// DoRequest performs an HTTP request based on the given APIRequest.
func (c *ConfluxAPIClient) DoRequest(ctx context.Context, req *models.APIRequest) (*models.APIResponse, error) {
	c.logger.Info("Starting HTTP request", "method", req.Method, "url", req.URL)

	// Create the HTTP request using the provided method, URL, and body.
	httpReq, err := http.NewRequestWithContext(ctx, req.Method, req.URL, req.Body)
	if err != nil {
		c.logger.Error("Failed to create HTTP request", "error", err)
		return nil, err
	}

	// Set any custom headers.
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	var token models.TokenResponse

	err = c.cache.GetMulti(ctx, []string{c.Code, "Token"}, &token)

	// If a BearerToken is set, add it to the request.
	if token.AccessToken != "" {
		c.logger.Debug("Adding bearer token to request")
		httpReq.Header.Set("Authorization", "Bearer " + token.AccessToken)
	}

	// Execute the HTTP request.
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		c.logger.Error("Failed to execute HTTP request", "error", err)
		return nil, err
	}
	defer resp.Body.Close()

	// Read the response body.
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		c.logger.Error("Failed to read response body", "error", err)
		return nil, err
	}

	c.logger.Info("HTTP request completed successfully", "statusCode", resp.StatusCode)

	return &models.APIResponse{
		StatusCode: resp.StatusCode,
		Body:       bodyBytes,
	}, nil
}

// FetchTokens delegates token fetching to the authentication strategy.
func (c *ConfluxAPIClient) FetchTokens(ctx context.Context, apiName string) (*models.TokenResponse, error) {
	c.logger.Info("Fetching tokens", "apiName", apiName)
	tokens, err := c.auth.FetchTokens(ctx, apiName)
	if err != nil {
		c.logger.Error("Failed to fetch tokens", "apiName", apiName, "error", err)
		return nil, err
	}
	c.logger.Info("Successfully fetched tokens", "apiName", apiName)
	return tokens, nil
}

// RefreshTokens delegates token refreshing to the authentication strategy.
func (c *ConfluxAPIClient) RefreshTokens(ctx context.Context, apiName string) (*models.TokenResponse, error) {
	c.logger.Info("Refreshing tokens", "apiName", apiName)
	tokens, err := c.auth.RefreshTokens(ctx, apiName)
	if err != nil {
		c.logger.Error("Failed to refresh tokens", "apiName", apiName, "error", err)
		return nil, err
	}
	c.logger.Info("Successfully refreshed tokens", "apiName", apiName)
	return tokens, nil
}
