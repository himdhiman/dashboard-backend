package limiter

import (
	"context"
	"time"

	"github.com/himdhiman/dashboard-backend/libs/cache"
	"github.com/himdhiman/dashboard-backend/libs/logger"
	"github.com/himdhiman/dashboard-backend/libs/mongo"
)

type ConfigSyncWorker struct {
	mongoRepo   *mongo.Repository[RateLimitConfig]
	redisClient *cache.CacheClient
	logger      logger.Logger
}

func NewConfigSyncWorker(mongoDatabaseName, mongoCollectionName string, redisClient *cache.CacheClient, logger logger.Logger) *ConfigSyncWorker {
	repo := mongo.Repository[RateLimitConfig]{
		Collection: mongo.GetCollection(mongoDatabaseName, mongoCollectionName),
	}
	return &ConfigSyncWorker{mongoRepo: &repo, redisClient: redisClient, logger: logger}
}

func (w *ConfigSyncWorker) Start(ctx context.Context) {
	for {
		w.syncConfigs(ctx)
		time.Sleep(5 * time.Minute)
	}
}

func (w *ConfigSyncWorker) syncConfigs(ctx context.Context) {
	configs, err := w.mongoRepo.Find(ctx, "rate_limit_configs", nil)
	if err != nil {
		w.logger.Error("Failed to fetch rate limit configs:", err)
		return
	}

	for _, config := range configs {
		// Sync default config
		defaultKey := "rate_limit:" + config.ServiceName + ":default"
		_ = w.redisClient.Set(ctx, defaultKey, config.DefaultConfig, config.DefaultConfig.TimeWindow)

		// Sync endpoint-specific configs
		for _, endpoint := range config.EndpointConfigs {
			key := "rate_limit:" + config.ServiceName + ":" + endpoint.Endpoint
			_ = w.redisClient.Set(ctx, key, endpoint, endpoint.TimeWindow)
		}
	}
	w.logger.Info("Rate limit configs synced successfully.")
}
