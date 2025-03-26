package shipping

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/himdhiman/dashboard-backend/libs/constants"
	"github.com/himdhiman/dashboard-backend/libs/logger"
	"github.com/himdhiman/dashboard-backend/libs/mappers"
	"github.com/himdhiman/dashboard-backend/libs/mongo"
	"github.com/himdhiman/dashboard-backend/libs/mongo/repository"
	"github.com/himdhiman/dashboard-backend/libs/task"
	purchaseOrder_models "github.com/himdhiman/dashboard-backend/services/purchaseOrder-service/models"
	"github.com/himdhiman/dashboard-backend/services/shipping-service/config"
	"github.com/himdhiman/dashboard-backend/services/shipping-service/models"
	"github.com/himdhiman/dashboard-backend/services/shipping-service/routes"
	"github.com/himdhiman/dashboard-backend/services/shipping-service/services"
)

type ShippingServices struct {
	ShippingService *services.ShippingService
}

func InitializeShippingService(router *gin.Engine, ctx context.Context, config *config.ShippingServiceConfig,
	logger logger.ILogger, mongoClient mongo.IMongoClient, taskManager *task.TaskManager) (*ShippingServices, error) {

	ShippingMarkCollection, err := mongoClient.GetCollection(context.Background(), constants.ShippingMarkCollection)
	if err != nil {
		logger.Fatal("Failed to connect to Collection", "error", err)
		return nil, err
	}

	shippingMarkRepo := repository.Repository[models.ShippingMark]{Collection: ShippingMarkCollection}

	ShippingProviderCollection, err := mongoClient.GetCollection(context.Background(), constants.ShippingProviderCollection)
	if err != nil {
		logger.Fatal("Failed to connect to Collection", "error", err)
		return nil, err
	}

	shippingProviderRepo := repository.Repository[models.ShippingProvider]{Collection: ShippingProviderCollection}

	purchaseOrderProductsCollection, err := mongoClient.GetCollection(context.Background(), constants.PurchaseOrderProductCollection)
	if err != nil {
		logger.Fatal("Failed to connect to Collection", "error", err)
		return nil, err
	}

	purchaseOrderProductsRepo := repository.Repository[purchaseOrder_models.PurchaseOrderProducts]{Collection: purchaseOrderProductsCollection}

	mapper := mappers.NewMapper()

	shippingService := services.NewShippingService(logger, mapper, *config.ProductService, taskManager, &shippingMarkRepo, &shippingProviderRepo, &purchaseOrderProductsRepo)

	routes.AddShippingRoutes(router, logger, &routes.ShippingService{
		ShippingService: shippingService,
	})

	return &ShippingServices{
		ShippingService: shippingService,
	}, nil
}
