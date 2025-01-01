package controllers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/himdhiman/dashboard-backend/libs/logger"
	"github.com/himdhiman/dashboard-backend/libs/task"
	"github.com/himdhiman/dashboard-backend/services/sentinel-service/services"
)

type UnicommerceController struct {
	Logger      logger.ILogger
	Service     *services.UnicommerceService
	TaskManager *task.TaskManager
}

func NewUnicommerceController(logger logger.ILogger, service *services.UnicommerceService, taskManager *task.TaskManager) *UnicommerceController {
	return &UnicommerceController{
		Logger:      logger,
		Service:     service,
		TaskManager: taskManager,
	}
}

func (uc *UnicommerceController) GetProducts(c *gin.Context) {
	// Parse query parameters
	pageNumberStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")
	skuCode := c.DefaultQuery("skuCode", "")

	pageNumber, err := strconv.Atoi(pageNumberStr)
	if err != nil || pageNumber < 1 {
		pageNumber = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 10
	}

	/// Fetch products
	ctx := c.Request.Context()
	products, err := uc.Service.GetProducts(ctx, skuCode, pageNumber, limit)
	if err != nil {
		uc.Logger.Error("Error fetching products", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch products"})
		return
	}

	c.JSON(http.StatusOK, products)
}

// FetchProducts fetches products from the Unicommerce API, runs the task in the background and returns the task ID
func (uc *UnicommerceController) FetchProducts(c *gin.Context) {
	fetchProductsTask := func(params map[string]interface{}) {
		ctx := context.Background()
		err := uc.Service.FetchProducts(ctx)
		if err != nil {
			uc.Logger.Error("Error fetching products", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch products"})
			return
		}
	}

	taskID, err := uc.TaskManager.RunTask("FetchProducts", nil, fetchProductsTask)
	if err != nil {
		uc.Logger.Error("Error running fetch products task", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to run fetch products task"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"task_id": taskID})
}
