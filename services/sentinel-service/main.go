package main

import (
	"context"
	"net/http"

	"github.com/himdhiman/dashboard-backend/libs/cache"
	"github.com/himdhiman/dashboard-backend/libs/crypto"
	"github.com/himdhiman/dashboard-backend/libs/logger"
	"github.com/himdhiman/dashboard-backend/libs/mongo"
	"github.com/himdhiman/dashboard-backend/libs/task"

	"github.com/himdhiman/dashboard-backend/services/sentinel-service/auth"
	"github.com/himdhiman/dashboard-backend/services/sentinel-service/constants"
	"github.com/himdhiman/dashboard-backend/services/sentinel-service/routes"
	"github.com/himdhiman/dashboard-backend/services/sentinel-service/schedulers"
	"github.com/himdhiman/dashboard-backend/services/sentinel-service/services"
	"github.com/himdhiman/dashboard-backend/services/sentinel-service/worker"
)

func main() {

	ctx := context.Background()
	logger := logger.New(logger.DefaultConfig("Sentinel-Service")).WithContext(ctx)

	mongoConfig := mongo.NewMongoConfig("mongodb://localhost:27017", "Dashboard")
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

	cacheConfig := cache.NewCacheConfig("localhost", 6379, "", 0, 0, "sentinel")
	cache, err := cache.NewCacheClient(cacheConfig, logger)
	if err != nil {
		logger.Error("Failed to connect to Redis", "error", err)
		return
	}

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
	tokenManager := auth.NewTokenManager(cache, logger, cryptoInstance, constants.UNICOM_API_CODE, authentication)

	collectionName = "unicom_products"
	collection, err = mongoClient.GetCollection(context.Background(), collectionName)
	if err != nil {
		logger.Fatal("Failed to connect to Collection", "error", err)
	}

	unicommerceService := services.NewUnicommerceService(tokenManager, logger, collection)

	taskCollectionName := "sentinel_tasks"
	collection, err = mongoClient.GetCollection(context.Background(), taskCollectionName)
	if err != nil {
		logger.Fatal("Failed to connect to Collection", "error", err)
	}
	taskManager := task.NewTaskManager(collection, logger)

	exportJobSchedulerCollectionName := "sentinel_schedulers"
	collection, err = mongoClient.GetCollection(context.Background(), exportJobSchedulerCollectionName)
	if err != nil {
		logger.Fatal("Failed to connect to Collection", "error", err)
	}
	exportJobScheduler := schedulers.NewExportJobScheduler(collection, unicommerceService, logger)
	exportJobScheduler.Start(ctx)

	// Set up router
	router := routes.SetupRouter(logger, unicommerceService, taskManager)

	// Start the server
	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatal("listen: ", err)
	}
}
