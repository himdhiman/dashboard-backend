package worker

import (
	"context"
	"time"

	"github.com/himdhiman/dashboard-backend/libs/cache"
	"github.com/himdhiman/dashboard-backend/libs/logger"
	mongo_models "github.com/himdhiman/dashboard-backend/libs/mongo/models"
	"github.com/himdhiman/dashboard-backend/libs/mongo/repository"

	"github.com/himdhiman/dashboard-backend/services/sentinel-service/auth"
	"github.com/himdhiman/dashboard-backend/services/sentinel-service/models"
)

func StartConfigSync(mongoCollection *mongo_models.MongoCollection, cache *cache.CacheClient, logger logger.LoggerInterface, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	mongoRepo := repository.Repository[models.APIConfig]{Collection: mongoCollection}

	for {
		select {
		case <-ticker.C:
			configSync(&mongoRepo, cache, logger)
		}
	}
}

func configSync(mongoRepo *repository.Repository[models.APIConfig], cache *cache.CacheClient, logger logger.LoggerInterface) {
	ctx := context.Background()
	filter := map[string]interface{}{}
	configs, err := mongoRepo.Find(ctx, filter)
	if err != nil {
		// log error
		return
	}

	for _, config := range configs {
		logger.Info("Config", "config", config)

		apiName := config.ApiName

		endpointKey := apiName + ":endpoint"
		rateLimitKey := apiName + ":rate_limit"
		authTypeKey := apiName + ":auth_type"

		cache.Set(ctx, endpointKey, config.Endpoint, 0)
		cache.Set(ctx, rateLimitKey, config.RateLimit, 0)
		cache.Set(ctx, authTypeKey, config.Authorization.Type, 0)

		if config.Authorization.Type == auth.BASIC_AUTH {
			usernameKey := apiName + ":username"
			passwordKey := apiName + ":password"

			cache.Set(ctx, usernameKey, config.Authorization.BasicAuth.Credentials.Username, 0)
			cache.Set(ctx, passwordKey, config.Authorization.BasicAuth.Credentials.Password, 0)
		}

	}
}
