package services

import (
	"github.com/himdhiman/dashboard-backend/libs/logger"
	mongo_models "github.com/himdhiman/dashboard-backend/libs/mongo/models"
	"github.com/himdhiman/dashboard-backend/libs/mongo/repository"
	"github.com/himdhiman/dashboard-backend/services/purchaseOrder-service/models"
)

type UnicommerceService struct {
	Logger                  logger.ILogger
	PurchaseOrderRepository *repository.Repository[models.PurchaseOrder]
}

func NewUnicommerceService(logger logger.ILogger,
	po_collections *mongo_models.MongoCollection) *UnicommerceService {

	purchaseOrderRepo := repository.Repository[models.PurchaseOrder]{Collection: po_collections}

	return &UnicommerceService{
		Logger:                  logger,
		PurchaseOrderRepository: &purchaseOrderRepo,
	}
}
