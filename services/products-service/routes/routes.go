package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/himdhiman/dashboard-backend/libs/logger"
	"github.com/himdhiman/dashboard-backend/services/products-service/controllers"
	"github.com/himdhiman/dashboard-backend/services/products-service/services"
)

type ProductsServices struct {
	UnicommerceProductsService *services.UnicommerceProductsService
	ProductsService            *services.ProductsService
}

func AddProductsRoutes(router *gin.Engine, logger logger.ILogger, productsServices *ProductsServices) *gin.Engine {

	unicommerceController := controllers.NewUnicommerceController(logger, productsServices.UnicommerceProductsService)
	productsController := controllers.NewProductsController(logger, productsServices.ProductsService)

	router.GET("/unicommerce/products", productsController.GetProducts)
	router.POST("/search-products", productsController.SearchProduct)
	router.GET("/product-bundles", productsController.GetProductBundles)

	router.POST("/unicommerce/create/job", unicommerceController.CreateExportJob)
	router.POST("/unicommerce/adjust/inventory", unicommerceController.AdjustUnicommerceInventory)

	return router
}
