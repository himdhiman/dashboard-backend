package purchaseOrder

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/himdhiman/dashboard-backend/libs/constants"
	"github.com/himdhiman/dashboard-backend/libs/logger"
	"github.com/himdhiman/dashboard-backend/libs/mappers"
	"github.com/himdhiman/dashboard-backend/libs/mongo"
	"github.com/himdhiman/dashboard-backend/libs/mongo/repository"
	"github.com/himdhiman/dashboard-backend/libs/task"
	product_services "github.com/himdhiman/dashboard-backend/services/products-service/services"
	"github.com/himdhiman/dashboard-backend/services/purchaseOrder-service/config"
	"github.com/himdhiman/dashboard-backend/services/purchaseOrder-service/models"
	"github.com/himdhiman/dashboard-backend/services/purchaseOrder-service/routes"
	"github.com/himdhiman/dashboard-backend/services/purchaseOrder-service/services"
)

type PurchaseOrderServices struct {
	PurchaseOrderService *services.PurchaseOrderService
}

func InitializePurchaseOrderService(router *gin.Engine, ctx context.Context, config *config.PurchaseOrderServiceConfig, logger logger.ILogger, mongoClient mongo.IMongoClient, taskManager *task.TaskManager, productsService product_services.ProductsService) (*PurchaseOrderServices, error) {

	purchaseOrderCollection, err := mongoClient.GetCollection(context.Background(), constants.PurchaseOrderCollection)
	if err != nil {
		logger.Fatal("Failed to connect to Collection", "error", err)
		return nil, err
	}

	purchaseOrderRepo := repository.Repository[models.PurchaseOrder]{Collection: purchaseOrderCollection}

	purchaseOrderProductsCollection, err := mongoClient.GetCollection(context.Background(), constants.PurchaseOrderProductCollection)
	if err != nil {
		logger.Fatal("Failed to connect to Collection", "error", err)
		return nil, err
	}

	purchaseOrderProductsRepo := repository.Repository[models.PurchaseOrderProducts]{Collection: purchaseOrderProductsCollection}

	mapper := mappers.NewMapper()

	purchaseOrderService := services.NewPurchaseOrderService(logger, mapper, taskManager, &purchaseOrderRepo, &purchaseOrderProductsRepo, productsService)

	routes.AddPurchaseOrderRoutes(router, logger, &routes.PurchaseOrderService{
		PurchaseOrderService: purchaseOrderService,
	})

	return &PurchaseOrderServices{
		PurchaseOrderService: purchaseOrderService,
	}, nil
}
