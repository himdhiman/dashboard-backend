package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/himdhiman/dashboard-backend/libs/logger"
	"github.com/himdhiman/dashboard-backend/services/purchaseOrder-service/controllers"
	"github.com/himdhiman/dashboard-backend/services/purchaseOrder-service/services"
)

type PurchaseOrderService struct {
	PurchaseOrderService *services.PurchaseOrderService
}

func AddPurchaseOrderRoutes(router *gin.Engine, logger logger.ILogger, purchaseOrderService *PurchaseOrderService) *gin.Engine {

	purchaseOrderController := controllers.NewPurchaseOrderController(logger, purchaseOrderService.PurchaseOrderService)

	router.GET("/purchase-order", purchaseOrderController.GetPurchaseOrders)
	router.POST("/purchase-order", purchaseOrderController.CreatePurchaseOrder)
	router.PUT("/purchase-orders", purchaseOrderController.UpdatePurchaseOrder)

	return router
}
