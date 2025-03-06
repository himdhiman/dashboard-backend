package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/himdhiman/dashboard-backend/libs/logger"
	"github.com/himdhiman/dashboard-backend/services/products-service/controllers"
	"github.com/himdhiman/dashboard-backend/services/products-service/services"
)

func AddProductsRoutes(router *gin.Engine, logger logger.ILogger, unicommerceProductsService *services.UnicommerceProductsService, productsService *services.ProductsService) *gin.Engine {

	unicommerceController := controllers.NewUnicommerceController(logger, unicommerceProductsService)
	productsController := controllers.NewProductsController(logger, productsService)

	router.GET("/unicommerce/products", productsController.GetProducts)
	router.POST("/search-products", productsController.SearchProduct)

	router.POST("/unicommerce/create/job", unicommerceController.CreateExportJob)

	return router
}
