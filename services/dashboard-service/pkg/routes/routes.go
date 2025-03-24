package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/himdhiman/dashboard-backend/libs/logger"
	"github.com/himdhiman/dashboard-backend/libs/task"
	"github.com/himdhiman/dashboard-backend/services/dashboard-service/pkg/controllers"
	"github.com/himdhiman/dashboard-backend/services/dashboard-service/pkg/middlewares"
)

func SetupRouter(logger logger.ILogger, taskManager *task.TaskManager) *gin.Engine {
	router := gin.Default()

	// Add CORS middleware
	router.Use(middlewares.CORSMiddleware())
	router.Use(middlewares.CorrelationIDMiddleware())

	controller := controllers.NewController(logger, taskManager)

	router.GET("/health", controller.HealthCheck)
	router.GET("/tasks/:task_id", controller.GetTaskStatus)

	return router
}
