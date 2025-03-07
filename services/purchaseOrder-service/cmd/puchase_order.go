package purchaseOrder

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/himdhiman/dashboard-backend/libs/logger"
	"github.com/himdhiman/dashboard-backend/libs/mongo"
	"github.com/himdhiman/dashboard-backend/libs/mongo/repository"
	"github.com/himdhiman/dashboard-backend/services/purchaseOrder-service/config"
	"github.com/himdhiman/dashboard-backend/services/purchaseOrder-service/models"
	"github.com/himdhiman/dashboard-backend/services/purchaseOrder-service/routes"
	"github.com/himdhiman/dashboard-backend/services/purchaseOrder-service/services"
)

func InitializePurchaseOrderService(router *gin.Engine, ctx context.Context, config *config.PurchaseOrderServiceConfig, logger logger.ILogger, mongoClient mongo.IMongoClient) error {

	collection, err := mongoClient.GetCollection(context.Background(), "PurchaseOrders")
	if err != nil {
		logger.Fatal("Failed to connect to Collection", "error", err)
		return err
	}

	purchaseOrderRepo := repository.Repository[models.PurchaseOrder]{Collection: collection}

	purchaseOrderService := services.NewPurchaseOrderService(logger, &purchaseOrderRepo)

	routes.AddPurchaseOrderRoutes(router, logger, &routes.PurchaseOrderService{
		PurchaseOrderService: purchaseOrderService,
	})

	return nil
}
