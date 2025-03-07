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
	"golang.org/x/oauth2/google"

	products "github.com/himdhiman/dashboard-backend/services/products-service/cmd"
	products_config "github.com/himdhiman/dashboard-backend/services/products-service/config"
	purchaseOrder "github.com/himdhiman/dashboard-backend/services/purchaseOrder-service/cmd"
	purchaseOrder_config "github.com/himdhiman/dashboard-backend/services/purchaseOrder-service/config"
	"github.com/himdhiman/dashboard-backend/services/dashboard-service/config"
	"github.com/himdhiman/dashboard-backend/services/dashboard-service/routes"
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

	cryptoInstance := crypto.NewCrypto(projectConfig.SecretKey, projectConfig.InitializationVector)

	// Initialize Conflux service
	confluxService := conflux.NewConfluxService("Sentinel Service", &cache, logger, cryptoInstance, mongoClient)

	// Initialize Unicommerce service

	// Initialize task manager
	taskCollectionName := "sentinel_tasks"
	collection, err := mongoClient.GetCollection(context.Background(), taskCollectionName)
	if err != nil {
		logger.Fatal("Failed to connect to Collection", "error", err)
	}

	taskManager := task.NewTaskManager(collection, logger)

	// Set up router
	router := routes.SetupRouter(logger, taskManager)

	productsServiceConfig := products_config.ProductsServiceConfig{
		SpreadsheetID: projectConfig.SpreadsheetID,
		SheetName:     projectConfig.SheetName,
		Credentials:   nil,
	}

	// Load Google credentials from file
	credBytes, err := os.ReadFile(projectConfig.GoogleCredentialsPath)
	if err != nil {
		logger.Fatal("Failed to read credentials file", "error", err)
	}

	creds, err := google.CredentialsFromJSON(ctx, credBytes, "https://www.googleapis.com/auth/spreadsheets")
	if err != nil {
		logger.Fatal("Failed to parse credentials", "error", err)
	}

	productsServiceConfig.Credentials = creds

	err = products.InitializeProductsService(router, ctx, &productsServiceConfig, logger, cache, confluxService, mongoClient)
	if err != nil {
		logger.Fatal("Failed to initialize products service", "error", err)
	}

	purchaseOrderServiceConfig := purchaseOrder_config.PurchaseOrderServiceConfig{}
	err = purchaseOrder.InitializePurchaseOrderService(router, ctx, &purchaseOrderServiceConfig, logger, mongoClient)
	if err != nil {
		logger.Fatal("Failed to initialize purchase order service", "error", err)
	}

	// Start the server
	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatal("listen: ", err)
	}
}
