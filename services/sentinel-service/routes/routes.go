package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/himdhiman/dashboard-backend/libs/logger"
	"github.com/himdhiman/dashboard-backend/libs/task"
	"github.com/himdhiman/dashboard-backend/services/sentinel-service/controllers"
	"github.com/himdhiman/dashboard-backend/services/sentinel-service/services"
)

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func SetupRouter(logger logger.ILogger, unicommerceService *services.UnicommerceService, taskManager *task.TaskManager) *gin.Engine {
	router := gin.Default()

	// Add CORS middleware
	router.Use(CORSMiddleware())

	controller := controllers.NewController(logger, taskManager)
	router.GET("/tasks/:task_id", controller.GetTaskStatus)

	unicommerceController := controllers.NewUnicommerceController(logger, unicommerceService, taskManager)

	router.GET("/unicommerce/products", unicommerceController.GetProducts)
	// router.POST("/unicommerce/products/fetch", unicommerceController.FetchProducts)

	router.POST("/unicommerce/create/job", unicommerceController.CreateExportJob)

	router.POST("/search-products", unicommerceController.SearchProduct)

	router.GET("/unicommerce/purchase-order/get", unicommerceController.GetPurchaseOrders)
	router.POST("/unicommerce/purchase-order/create", unicommerceController.CreatePurchaseOrder)

	return router
}
