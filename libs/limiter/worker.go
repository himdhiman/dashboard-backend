// internal/limiter/worker.go
package limiter

import (
	"context"
	"time"

	"github.com/himdhiman/dashboard-backend/libs/logger"
	"github.com/himdhiman/dashboard-backend/libs/mongo"
	"github.com/himdhiman/dashboard-backend/libs/redis_cache"
)

type ConfigSyncWorker struct {
	mongoClient *mongo.Client
	redisClient *redis_cache.Client
	logger      logger.Logger
}

func NewConfigSyncWorker(mongoClient *mongo.Client, redisClient *redis_cache.Client, logger logger.Logger) *ConfigSyncWorker {
	return &ConfigSyncWorker{mongoClient: mongoClient, redisClient: redisClient, logger: logger}
}

func (w *ConfigSyncWorker) Start(ctx context.Context) {
	for {
		w.syncConfigs(ctx)
		time.Sleep(5 * time.Minute)
	}
}

func (w *ConfigSyncWorker) syncConfigs(ctx context.Context) {
	var configs []RateLimitConfig
	err := w.mongoClient.Find(ctx, "rate_limit_configs", nil, &configs)
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
