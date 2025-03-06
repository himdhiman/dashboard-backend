package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/himdhiman/dashboard-backend/libs/cache"
	conflux "github.com/himdhiman/dashboard-backend/libs/conflux/cmd"
	"github.com/himdhiman/dashboard-backend/libs/crypto"
	"github.com/himdhiman/dashboard-backend/libs/logger"
	"github.com/himdhiman/dashboard-backend/libs/mongo"
	"github.com/himdhiman/dashboard-backend/libs/task"
	"github.com/joho/godotenv"

	"github.com/himdhiman/dashboard-backend/services/sentinel-service/config"
	"github.com/himdhiman/dashboard-backend/services/sentinel-service/constants"
	"github.com/himdhiman/dashboard-backend/services/sentinel-service/routes"
	"github.com/himdhiman/dashboard-backend/services/sentinel-service/services"
	"github.com/himdhiman/dashboard-backend/services/sentinel-service/worker"
)


func main() {
	ctx := context.Background()
	logger := logger.New(logger.DefaultConfig("Sentinel-Service")).WithContext(ctx)

	// Load configuration
	projectConfig, err := config.LoadConfig()
	if err != nil {
		logger.Fatal("Failed to load configuration", "error", err)
		return
	}

	// Initialize MongoDB connection
	mongoConnectString := fmt.Sprintf("%s:%s@%s", projectConfig.MongoUser, projectConfig.MongoPassword, projectConfig.MongoHost)
	mongoConfig := mongo.NewMongoConfig(fmt.Sprintf("mongodb://%s:27017", mongoConnectString), "Dashboard")
	mongoClient, err := mongo.NewMongoClient(mongoConfig, logger)
	if err != nil {
		logger.Error("Failed to connect to MongoDB", "error", err)
		return
	}

	defer mongoClient.Disconnect(ctx)

	// Create cache configuration
	cacheConfig := &cache.CacheConfig{
		Prefix:  "sentinel",
		Timeout: 0,
	}

	cache, err := cache.NewMemoryCache(cacheConfig, logger)
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

	worker.StartConfigSync(collection, cache, logger)

	cryptoInstance := crypto.NewCrypto(projectConfig.SecretKey, projectConfig.InitializationVector)

	// Initialize collections
	collectionName = "unicom_purchase_orders"
	po_collection, err := mongoClient.GetCollection(context.Background(), collectionName)
	if err != nil {
		logger.Fatal("Failed to connect to Collection", "error", err)
	}


	// Initialize Conflux service
	confluxService := conflux.NewConfluxService(constants.UNICOM_API_CODE, &cache, logger, cryptoInstance, mongoClient)
	

	// Initialize Unicommerce service
	unicommerceService := services.NewUnicommerceService(unicommerceApiClient, logger, cache, collection, po_collection)

	// Initialize task manager
	taskCollectionName := "sentinel_tasks"
	collection, err = mongoClient.GetCollection(context.Background(), taskCollectionName)
	if err != nil {
		logger.Fatal("Failed to connect to Collection", "error", err)
	}

	taskManager := task.NewTaskManager(collection, logger)

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
