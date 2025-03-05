package client

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/himdhiman/dashboard-backend/libs/cache"
	"github.com/himdhiman/dashboard-backend/libs/conflux/pkg/auth"
	"github.com/himdhiman/dashboard-backend/libs/conflux/pkg/models"
	"github.com/himdhiman/dashboard-backend/libs/logger"
)

type ConfluxAPIClient struct {
	models.APIConfig
	tokenManager *auth.TokenManager
	logger       logger.ILogger
	cache        cache.Cacher
	httpClient   *http.Client
}

// NewConfluxAPIClient creates a new instance of ConfluxAPIClient
func NewConfluxAPIClient(config models.APIConfig, tokenManager *auth.TokenManager, logger logger.ILogger, cache cache.Cacher, httpClient *http.Client) *ConfluxAPIClient {
	return &ConfluxAPIClient{
		APIConfig:    config,
		tokenManager: tokenManager,
		logger:       logger,
		cache:        cache,
		httpClient:   httpClient,
	}
}

// GetBaseURL returns the configured BaseURL.
func (c *ConfluxAPIClient) GetBaseURL() string {
	return c.APIConfig.BaseURL
}

// DoRequest performs an HTTP request based on the given APIRequest.
func (c *ConfluxAPIClient) DoRequest(ctx context.Context, req *models.APIRequest) (*models.APIResponse, error) {
	// get the endpoint from the list of endpoints whose code matches with the api code
	var endpoint models.Endpoints
	for _, e := range c.APIConfig.Endpoints {
		if e.Code == req.ApiCode {
			endpoint = e
			break
		}
	}

	if endpoint.Code == "" {
		return nil, fmt.Errorf("endpoint not found for api code: %s", req.ApiCode)
	}

	endpointURL := c.APIConfig.BaseURL + endpoint.Path

	c.logger.Info("Starting HTTP request", "method", endpoint.Method, "url", endpoint.Code)

	// Create the HTTP request using the provided method, URL, and body.
	httpReq, err := http.NewRequestWithContext(ctx, string(endpoint.Method), endpointURL, req.Body)
	if err != nil {
		c.logger.Error("Failed to create HTTP request", "error", err)
		return nil, err
	}

	// Set any custom headers.
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	c.tokenManager.AuthenticateRequest(ctx, httpReq)

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
