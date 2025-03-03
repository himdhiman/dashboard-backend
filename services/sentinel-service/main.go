package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/himdhiman/dashboard-backend/libs/cache"
	"github.com/himdhiman/dashboard-backend/libs/crypto"
	"github.com/himdhiman/dashboard-backend/libs/logger"
	"github.com/himdhiman/dashboard-backend/libs/mongo"
	"github.com/himdhiman/dashboard-backend/libs/task"
	"github.com/himdhiman/dashboard-backend/libs/conflux"
	"github.com/joho/godotenv"

	"github.com/himdhiman/dashboard-backend/services/sentinel-service/routes"
	"github.com/himdhiman/dashboard-backend/services/sentinel-service/schedulers"
	"github.com/himdhiman/dashboard-backend/services/sentinel-service/services"
	"github.com/himdhiman/dashboard-backend/services/sentinel-service/worker"
)

func loadConfig(logger logger.ILogger) {
	// Check if we're in production
	appEnv := os.Getenv("APP_ENV")
	if appEnv != "production" {
		// Load from .env file in development
		if err := godotenv.Load(".env"); err != nil {
			logger.Warn("Warning: .env file not found, falling back to environment variables")
		}
	}
	// In production, will use environment variables directly
}

func main() {
	ctx := context.Background()
	logger := logger.New(logger.DefaultConfig("Sentinel-Service")).WithContext(ctx)

	loadConfig(logger)

	// Get MongoDB credentials from environment variables
	mongoUser := os.Getenv("MONGO_USER")
	mongoPass := os.Getenv("MONGO_PASSWORD")
	mongoHost := os.Getenv("MONGO_HOST")

	if mongoUser == "" || mongoPass == "" || mongoHost == "" {
		logger.Fatal("Required environment variables MONGO_USER, MONGO_PASSWORD, MONGO_HOST not set")
		return
	}

	// Get secret key and initialization vector for encryption
	secretKey := os.Getenv("SECRET_KEY")
	initializationVector := os.Getenv("INITIALIZATION_VECTOR")

	if secretKey == "" || initializationVector == "" {
		logger.Fatal("Required environment variables SECRET_KEY, INITIALIZATION_VECTOR not set")
		return
	}

	// Get Google Sheets credentials

	spreadsheetID := os.Getenv("SPREADSHEET_ID")
	sheetName := os.Getenv("SHEET_NAME")
	credentials := os.Getenv("GOOGLE_CREDENTIALS_PATH")

	if spreadsheetID == "" || sheetName == "" || credentials == "" {
		logger.Fatal("Required environment variables SPREADSHEET_ID, SHEET_NAME, GOOGLE_CREDENTIALS_PATH not set")
		return
	}

	mongoConnectString := fmt.Sprintf("%s:%s@%s", mongoUser, mongoPass, mongoHost)
	mongoConfig := mongo.NewMongoConfig(fmt.Sprintf("mongodb://%s:27017", mongoConnectString), "Dashboard")
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

	googleSheetsService := services.NewGoogleSheetsService(spreadsheetID, sheetName, credentials, logger)
	cryptoInstance := crypto.NewCrypto(secretKey, initializationVector)

	collectionName = "unicom_products"
	collection, err = mongoClient.GetCollection(context.Background(), collectionName)
	if err != nil {
		logger.Fatal("Failed to connect to Collection", "error", err)
	}

	collectionName = "unicom_purchase_orders"
	po_collection, err := mongoClient.GetCollection(context.Background(), collectionName)
	if err != nil {
		logger.Fatal("Failed to connect to Collection", "error", err)
	}

	confluxService := conflux.NewConfluxService(constants.UNICOM_API_CODE, cache, logger, cryptoInstance, collection)
	unicommerceApiClient := conflux.NewConfluxAPIClient(constants.UNICOM_API_CODE, tokenManager, http.DefaultClient, logger, cache)



	unicommerceService := services.NewUnicommerceService(tokenManager, googleSheetsService, logger, collection, po_collection)

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

	// start invetory snapshot scheduler
	inventorySnapShotScheduler := schedulers.NewInventorySnapShotScheduler(collection, unicommerceService, logger)
	inventorySnapShotScheduler.Start(ctx)

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
