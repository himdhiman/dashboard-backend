package config

import (
	"github.com/himdhiman/dashboard-backend/libs/cache"
	"github.com/himdhiman/dashboard-backend/libs/mongo"
)

type Config struct {
	MongoDB                    *mongo.MongoClient
	RedisCache                 *cache.CacheClient
	RateLimitterDatabaseName   string
	RateLimitterCollectionName string
	RateLimit                  int // in requests per minute
	WorkerSyncInterval         int // in seconds

}
