package interfaces

import (
	"context"

	"github.com/himdhiman/dashboard-backend/libs/conflux/pkg/models"
)

type AuthenticationStrategy interface {
	FetchTokens(ctx context.Context, apiName string) (*models.TokenResponse, error)
	RefreshTokens(ctx context.Context, apiName string) (*models.TokenResponse, error)
}

type APIClient interface {
	models.ApiClientConfig
	AuthenticationStrategy
	GetBaseURL() string
	DoRequest(ctx context.Context, req *models.APIRequest) (*models.APIResponse, error)
}
