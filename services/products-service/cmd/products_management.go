package products

import (
	"context"

	"github.com/himdhiman/dashboard-backend/libs/logger"
	"github.com/himdhiman/dashboard-backend/libs/mongo"
	"github.com/himdhiman/dashboard-backend/services/products-service/constants"
	"github.com/himdhiman/dashboard-backend/services/products-service/services"
)

func InitializeProductsManagementService(ctx context.Context, logger logger.ILogger, mongoClient mongo.IMongoClient) (*services.ManagementService, error) {

	collection, err := mongoClient.GetCollection(context.Background(), constants.UnicommerceProductsCollection)
	if err != nil {
		logger.Fatal("Failed to connect to Collection", "error", err)
		return nil, err
	}

	managementService := services.NewManagementService(logger, collection)

	return managementService, nil
}
