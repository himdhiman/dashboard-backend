package worker

import (
	"context"
	"time"

	"github.com/himdhiman/dashboard-backend/libs/cache"
	"github.com/himdhiman/dashboard-backend/libs/logger"
	mongo_models "github.com/himdhiman/dashboard-backend/libs/mongo/models"
	"github.com/himdhiman/dashboard-backend/libs/mongo/repository"

	"github.com/himdhiman/dashboard-backend/services/sentinel-service/constants"
	"github.com/himdhiman/dashboard-backend/services/sentinel-service/models"
)

func StartConfigSync(mongoCollection *mongo_models.MongoCollection, cache cache.Cacher, logger logger.ILogger) {
	mongoRepo := repository.Repository[models.APIConfig]{Collection: mongoCollection}

	// Run configSync immediately
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	configSync(ctx, &mongoRepo, cache, logger)
	cancel()
}

func configSync(ctx context.Context, mongoRepo repository.IRepository[models.APIConfig], cache cache.Cacher, logger logger.ILogger) {
	filter := map[string]interface{}{}
	configs, err := mongoRepo.Find(ctx, filter)
	if err != nil {
		logger.Error("Error fetching configs", "error", err)
		return
	}

	for _, config := range configs {
		logger.Info("Config", "config", config)

		apiCode := config.Code

		baseURLKey := constants.GetBaseURLKey(apiCode)
		authPathKey := constants.GetAuthPathKey(apiCode)
		authTypeKey := constants.GetAuthTypeKey(apiCode)
		authCredentialsKey := constants.GetAuthCredentialsKey(apiCode)

		cache.Set(ctx, baseURLKey, config.BaseURL)
		cache.Set(ctx, authPathKey, config.Authorization.Path)
		cache.Set(ctx, authTypeKey, config.Authorization.Type)
		cache.Set(ctx, authCredentialsKey, config.Authorization.Credentials)

		for _, endpoint := range config.Endpoints {
			apiPathKey := constants.GetApiPathKey(apiCode, endpoint.Code)
			apiMethodKey := constants.GetApiMethodKey(apiCode, endpoint.Code)
			apiRateLimitKey := constants.GetApiRateLimitKey(apiCode, endpoint.Code)
			apiTimeoutKey := constants.GetApiTimeoutKey(apiCode, endpoint.Code)

			cache.Set(ctx, apiPathKey, endpoint.Path)
			cache.Set(ctx, apiMethodKey, endpoint.Method)
			cache.Set(ctx, apiRateLimitKey, endpoint.RateLimit)
			cache.Set(ctx, apiTimeoutKey, endpoint.Timeout)
		}
	}
}
