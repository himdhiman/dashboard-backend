package main

import (
	"context"

	"github.com/himdhiman/dashboard-backend/libs/logger"
	"github.com/himdhiman/dashboard-backend/services/management-service/pkg/config"
	"github.com/himdhiman/dashboard-backend/services/management-service/pkg/utils"
	products "github.com/himdhiman/dashboard-backend/services/products-service/cmd"
)

const (
	appName = "Management-Service"
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

	productManagementService, err := products.InitializeProductsManagementService(ctx, logger, mongoClient)

	productManagementService.UpdateLastProcuredRMBPriceFromCSV(ctx, "/Users/himanshudhiman/Downloads/LastRMB.csv")

}
