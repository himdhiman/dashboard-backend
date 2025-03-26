package main

import (
	"context"
	"net/http"

	conflux "github.com/himdhiman/dashboard-backend/libs/conflux/cmd"
	"github.com/himdhiman/dashboard-backend/libs/constants"
	"github.com/himdhiman/dashboard-backend/libs/crypto"
	"github.com/himdhiman/dashboard-backend/libs/logger"
	"github.com/himdhiman/dashboard-backend/libs/task"

	"github.com/himdhiman/dashboard-backend/services/dashboard-service/pkg/config"
	"github.com/himdhiman/dashboard-backend/services/dashboard-service/pkg/routes"
	"github.com/himdhiman/dashboard-backend/services/dashboard-service/pkg/utils"
	products "github.com/himdhiman/dashboard-backend/services/products-service/cmd"
	products_config "github.com/himdhiman/dashboard-backend/services/products-service/config"
	purchaseOrder "github.com/himdhiman/dashboard-backend/services/purchaseOrder-service/cmd"
	shipping "github.com/himdhiman/dashboard-backend/services/shipping-service/cmd"
	shipping_config "github.com/himdhiman/dashboard-backend/services/shipping-service/config"

	purchaseOrder_config "github.com/himdhiman/dashboard-backend/services/purchaseOrder-service/config"
)

const (
	appName = "Dashboard-Service"
	port    = ":8080"
	envFile = "./configs/dev.env"
)

func main() {
	ctx := context.Background()
	logger := logger.New(logger.DefaultConfig(appName)).WithContext(ctx)

	// Load configuration
	projectConfig, err := config.LoadConfig(envFile)
	if err != nil {
		logger.Fatal("Failed to load configuration", "error", err)
		return
	}

	// Initialize MongoDB connection
	mongoClient, err := utils.GetMongoClient(ctx, projectConfig, logger)
	if err != nil {
		logger.Fatal("Failed to connect to MongoDB", "error", err)
		return
	}

	defer mongoClient.Disconnect(ctx)

	// Create cache configuration
	cache, err := utils.GetMemoryCache(ctx, projectConfig, logger)
	if err != nil {
		logger.Fatal("Failed to connect to Memory Cache", "error", err)
		return
	}

	defer cache.Close()

	// Initialize crypto instance
	cryptoInstance := crypto.NewCrypto(projectConfig.SecretKey, projectConfig.InitializationVector)

	// Initialize Conflux service
	confluxService := conflux.NewConfluxService("Dashboard Service", &cache, logger, cryptoInstance, mongoClient)

	// Initialize task manager
	collection, err := mongoClient.GetCollection(context.Background(), constants.TaskCollection)
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
	creds, err := utils.LoadGoogleCreds(ctx, projectConfig.GoogleCredentialsPath, logger)
	if err != nil {
		logger.Fatal("Failed to load Google credentials", "error", err)
		return
	}

	productsServiceConfig.Credentials = creds

	productsServices, err := products.InitializeProductsService(router, ctx, &productsServiceConfig, logger, cache, confluxService, mongoClient)
	if err != nil {
		logger.Fatal("Failed to initialize products service", "error", err)
	}

	shippingServiceConfig := shipping_config.ShippingServiceConfig{
		ProductService: productsServices.ProductsService,
	}

	shippingServices, err := shipping.InitializeShippingService(router, ctx, &shippingServiceConfig, logger, mongoClient, taskManager)
	if err != nil {
		logger.Fatal("Failed to initialize shipping service", "error", err)
	}

	purchaseOrderServiceConfig := purchaseOrder_config.PurchaseOrderServiceConfig{
		ProductService:  productsServices.ProductsService,
		ShippingService: shippingServices.ShippingService,
	}
	_, err = purchaseOrder.InitializePurchaseOrderService(router, ctx, &purchaseOrderServiceConfig, logger, mongoClient, taskManager)
	if err != nil {
		logger.Fatal("Failed to initialize purchase order service", "error", err)
	}

	// Start the server
	srv := &http.Server{
		Addr:    port,
		Handler: router,
	}

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatal("listen: ", err)
	}
}
