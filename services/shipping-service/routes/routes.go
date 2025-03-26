package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/himdhiman/dashboard-backend/libs/logger"
	"github.com/himdhiman/dashboard-backend/services/shipping-service/controllers"
	"github.com/himdhiman/dashboard-backend/services/shipping-service/services"
)

type ShippingService struct {
	ShippingService *services.ShippingService
}

func AddShippingRoutes(router *gin.Engine, logger logger.ILogger, shippingService *ShippingService) *gin.Engine {

	shippingController := controllers.NewShippingController(logger, shippingService.ShippingService)

	router.GET("/shipping-marks-list", shippingController.ListShippingMarks)

	router.GET("/shipping-marks", shippingController.GetShippingMark)
	router.PUT("/shipping-marks", shippingController.UpdateShippingMark)

	return router
}
