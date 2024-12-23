package auth

import (
	"context"

	"github.com/himdhiman/dashboard-backend/services/sentinel-service/models"
)

type AuthenticationStrategy interface {
	FetchTokens(ctx context.Context, apiName string) (*models.TokenResponse, error)
	RefreshTokens(ctx context.Context, apiName string) (*models.TokenResponse, error)
}
