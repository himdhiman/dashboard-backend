package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/himdhiman/dashboard-backend/libs/logger"
	"github.com/himdhiman/dashboard-backend/libs/task"
	"github.com/himdhiman/dashboard-backend/services/sentinel-service/controllers"
	"github.com/himdhiman/dashboard-backend/services/sentinel-service/services"
)

func SetupRouter(logger logger.ILogger, unicommerceService *services.UnicommerceService, taskManager *task.TaskManager) *gin.Engine {
	router := gin.Default()

	unicommerceController := controllers.NewUnicommerceController(logger, unicommerceService, taskManager)

	router.GET("/unicommerce/products", unicommerceController.GetProducts)
	router.POST("/unicommerce/products/fetch", unicommerceController.FetchProducts)

	return router
}
