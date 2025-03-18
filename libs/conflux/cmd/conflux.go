package conflux

import (
	"context"
	"net/http"
	"time"

	"github.com/himdhiman/dashboard-backend/libs/cache"
	"github.com/himdhiman/dashboard-backend/libs/conflux/pkg/auth"
	authstrategies "github.com/himdhiman/dashboard-backend/libs/conflux/pkg/auth/auth-strategies"
	"github.com/himdhiman/dashboard-backend/libs/conflux/pkg/client"
	conflux_errors "github.com/himdhiman/dashboard-backend/libs/conflux/pkg/error"
	interfaces "github.com/himdhiman/dashboard-backend/libs/conflux/pkg/interface"
	"github.com/himdhiman/dashboard-backend/libs/conflux/pkg/models"
	"github.com/himdhiman/dashboard-backend/libs/crypto"
	"github.com/himdhiman/dashboard-backend/libs/logger"
	"github.com/himdhiman/dashboard-backend/libs/mongo"
	"github.com/himdhiman/dashboard-backend/libs/mongo/repository"
	"github.com/himdhiman/dashboard-backend/libs/constants"
)

type AuthStrategy string

const (
	AuthStrategyBasic  AuthStrategy = "basic"
	AuthStrategyBearer AuthStrategy = "bearer"
)

type ConfluxService struct {
	serviceName         string
	ApiConfigRepository repository.Repository[models.APIConfig]
	cache               *cache.Cacher
	crypto              *crypto.Crypto
	logger              logger.ILogger
}

func NewConfluxService(serviceName string, cache *cache.Cacher, logger logger.ILogger, crypto *crypto.Crypto, mongoClient mongo.IMongoClient) *ConfluxService {
	collection, err := mongoClient.GetCollection(context.Background(), constants.ConfluxApisCollection)
	if err != nil {
		logger.Fatal("Failed to connect to Collection", "error", err)
	}

	apiConfigRepository := repository.Repository[models.APIConfig]{Collection: collection}

	return &ConfluxService{
		serviceName:         serviceName,
		cache:               cache,
		logger:              logger,
		crypto:              crypto,
		ApiConfigRepository: apiConfigRepository,
	}
}

// CreateApiClient creates and returns an API client based on the provided API code.
func (cs *ConfluxService) CreateApiClient(apiCode string, authStrategyType AuthStrategy) (*client.ConfluxAPIClient, error) {
	apiConfig, err := cs.ApiConfigRepository.FindOne(context.Background(), map[string]interface{}{"code": apiCode})
	if err != nil {
		return nil, err
	}

	if apiConfig == nil {
		return nil, conflux_errors.ErrAPIConfigNotFound
	}
	var authStrategy interfaces.AuthenticationStrategy

	switch authStrategyType {
	case AuthStrategyBasic:
		authStrategy = authstrategies.NewBasicAuthStrategy(apiCode, models.Credentials{
			Username:     apiConfig.Authorization.Credentials.Username,
			ClientID:     apiConfig.Authorization.Credentials.ClientID,
			ClientSecret: apiConfig.Authorization.Credentials.ClientSecret,
		}, apiConfig.BaseURL+apiConfig.Authorization.Path, cs.logger, *cs.cache, cs.crypto)
	case AuthStrategyBearer:
		authStrategy = authstrategies.NewBasicAuthStrategy(apiCode, models.Credentials{
			Username:     apiConfig.Authorization.Credentials.Username,
			ClientID:     apiConfig.Authorization.Credentials.ClientID,
			ClientSecret: apiConfig.Authorization.Credentials.ClientSecret,
		}, apiConfig.BaseURL+apiConfig.Authorization.Path, cs.logger, *cs.cache, cs.crypto)
	}

	tokenManager := auth.NewTokenManager(*cs.cache, cs.logger, cs.crypto, apiCode, authStrategy)

	httpClient := &http.Client{
		Timeout: time.Duration(10) * time.Second,
	}

	return client.NewConfluxAPIClient(*apiConfig, tokenManager, cs.logger, *cs.cache, httpClient), nil
}
