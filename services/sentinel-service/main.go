package main

import (
	"context"

	"github.com/himdhiman/dashboard-backend/libs/cache"
	"github.com/himdhiman/dashboard-backend/libs/crypto"
	"github.com/himdhiman/dashboard-backend/libs/logger"
	"github.com/himdhiman/dashboard-backend/libs/mongo"
	"github.com/himdhiman/dashboard-backend/libs/mongo/models"

	"github.com/himdhiman/dashboard-backend/services/sentinel-service/auth"
	"github.com/himdhiman/dashboard-backend/services/sentinel-service/worker"
)

func main() {

	logger := logger.New(logger.DefaultConfig())

	ctx := context.Background()

	mongoConfig := models.Config{
		MongoURL:     "mongodb://localhost:27017",
		DatabaseName: "Dashboard",
	}

	mongoClient, err := mongo.NewMongoClient(mongoConfig, logger)
	if err != nil {
		logger.Error("Failed to connect to MongoDB", "error", err)
		return
	}

	defer mongoClient.Disconnect(ctx)

	collectionName := "sentinel_apis"
	collection, err := mongoClient.GetCollection(context.Background(), collectionName)
	if err != nil {
		logger.Fatal("Failed to connect to Collection", "error", err)
	}

	cacheConfig := cache.CacheConfig{Host: "localhost", Port: 6379, Password: "", DB: 0, Timeout: 1, Prefix: "sentinel"}

	cache := cache.NewCacheClient(&cacheConfig, logger)

	ctx = context.Background()
	err = cache.Ping(ctx)
	if err != nil {
		logger.Error("Failed to connect to Redis", "error", err)
		return
	}

	worker.StartConfigSync(collection, cache, logger, 1000)

	secretKey := "Rvdf8NYhzKLpsRrMb7th34bW8bqh4HdT"
	initializationVector := "zUzT1iPfLMw80idf"

	cryptoInstance := crypto.NewCrypto(secretKey, initializationVector)

	authentication := auth.NewAuthentication(cache, logger, cryptoInstance)

	tokens, err := authentication.FetchTokens(ctx, "unicommerce")

	if err != nil {
		logger.Error("Failed to fetch tokens", "error", err)
		return
	}

	logger.Info("Tokens", "tokens", tokens)

}
