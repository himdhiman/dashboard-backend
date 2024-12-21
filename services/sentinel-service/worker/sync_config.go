package worker

import (
	"context"
	"time"

	"github.com/himdhiman/dashboard-backend/libs/cache"
	"github.com/himdhiman/dashboard-backend/libs/logger"
	mongo_models "github.com/himdhiman/dashboard-backend/libs/mongo/models"
	"github.com/himdhiman/dashboard-backend/libs/mongo/repository"

	"github.com/himdhiman/dashboard-backend/services/sentinel-service/models"
)

func StartConfigSync(mongoCollection *mongo_models.MongoCollection, cache *cache.CacheClient, logger logger.LoggerInterface, interval time.Duration) {
	ticker := time.NewTicker(interval * time.Second)
	defer ticker.Stop()

	mongoRepo := repository.Repository[models.APIConfig]{Collection: mongoCollection}

	// Run configSync immediately
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	configSync(ctx, &mongoRepo, cache, logger)
	cancel()

	go func() {
		for range ticker.C {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
            configSync(ctx, &mongoRepo, cache, logger)
            cancel()
		}
	}()
}

func configSync(ctx context.Context, mongoRepo *repository.Repository[models.APIConfig], cache *cache.CacheClient, logger logger.LoggerInterface) {
	filter := map[string]interface{}{}
	configs, err := mongoRepo.Find(ctx, filter)
	if err != nil {
		logger.Error("Error fetching configs", "error", err)
        return
	}

	for _, config := range configs {
		logger.Info("Config", "config", config)

		apiName := config.ApiName

		endpointKey := apiName + ":endpoint"
		pathKey := apiName + ":path"
		methodKey := apiName + ":method"
		rateLimitKey := apiName + ":rate_limit"
		authTypeKey := apiName + ":auth_type"

		cache.Set(ctx, endpointKey, config.Endpoint)
		cache.Set(ctx, pathKey, config.Path)
		cache.Set(ctx, methodKey, config.Method)
		cache.Set(ctx, rateLimitKey, config.RateLimit)
		cache.Set(ctx, authTypeKey, config.Authorization.Type)

		usernameKey := apiName + ":username"
		clientIdKey := apiName + ":client_id"
		clientSecretKey := apiName + ":client_secret"

		cache.Set(ctx, usernameKey, config.Authorization.OAuthConfig.Username)
		cache.Set(ctx, clientIdKey, config.Authorization.OAuthConfig.ClientID)
		cache.Set(ctx, clientSecretKey, config.Authorization.OAuthConfig.ClientSecret)

	}
}
