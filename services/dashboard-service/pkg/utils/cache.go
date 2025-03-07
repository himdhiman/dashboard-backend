package utils

import (
	"context"

	"github.com/himdhiman/dashboard-backend/libs/cache"
	"github.com/himdhiman/dashboard-backend/libs/logger"
	"github.com/himdhiman/dashboard-backend/services/dashboard-service/pkg/config"
)

func GetMemoryCache(ctx context.Context, projectConfig *config.ProjectConfig, logger logger.ILogger) (cache.Cacher, error) {
	cacheConfig := &cache.CacheConfig{
		Prefix:  "dashboard",
		Timeout: 0,
	}

	cache, err := cache.NewMemoryCache(cacheConfig, logger)
	if err != nil {
		logger.Error("Failed to connect to Memory Cache", "error", err)
		return nil, err
	}

	err = cache.Ping(ctx)
	if err != nil {
		logger.Error("Failed to connect to Memory Cache", "error", err)
		return nil, err
	}

	return cache, nil
}
