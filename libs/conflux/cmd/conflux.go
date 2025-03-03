package conflux

import (
	"context"
	"net/http"
	"time"

	"github.com/himdhiman/dashboard-backend/libs/cache"
	"github.com/himdhiman/dashboard-backend/libs/conflux/pkg/models"
	"github.com/himdhiman/dashboard-backend/libs/conflux/pkg/auth"
	"github.com/himdhiman/dashboard-backend/libs/conflux/pkg/client"
	interfaces "github.com/himdhiman/dashboard-backend/libs/conflux/pkg/interface"
	"github.com/himdhiman/dashboard-backend/libs/crypto"
	"github.com/himdhiman/dashboard-backend/libs/logger"
	mongo_models "github.com/himdhiman/dashboard-backend/libs/mongo/models"
	"github.com/himdhiman/dashboard-backend/libs/mongo/repository"
)

type ConfluxService struct {
	serviceName         string
	ApiConfigCollection repository.Repository[models.APIConfig]
	cache               *cache.Cacher
	crypto              *crypto.Crypto
	logger              logger.ILogger
}

func NewConfluxService(serviceName string, cache *cache.Cacher, logger logger.ILogger, crypto *crypto.Crypto, apiConfigCollection *mongo_models.MongoCollection) *ConfluxService {
	apiConfigRepository := repository.Repository[models.APIConfig]{Collection: apiConfigCollection}
	return &ConfluxService{
		serviceName:         serviceName,
		cache:               cache,
		logger:              logger,
		crypto:              crypto,
		ApiConfigCollection: apiConfigRepository,
	}
}

// CreateApiClient creates and returns an API client based on the provided API code.
func (cs *ConfluxService) CreateApiClient(apiCode string, authStrategy interfaces.AuthenticationStrategy) (*client.ConfluxAPIClient, error) {
	apiConfig, err := cs.ApiConfigCollection.FindOne(context.Background(), map[string]interface{}{"code": apiCode})
	if err != nil {
		return nil, err
	}

	tokenManager := auth.NewTokenManager(*cs.cache, cs.logger, cs.crypto, apiCode, authStrategy)

	httpClient := &http.Client{
		Timeout: time.Duration(5) * time.Second,
	}

	return client.NewConfluxAPIClient(*apiConfig, tokenManager, httpClient, cs.logger, *cs.cache), nil
}
