package controllers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/himdhiman/dashboard-backend/libs/logger"
	"github.com/himdhiman/dashboard-backend/libs/task"
	"github.com/himdhiman/dashboard-backend/services/sentinel-service/constants"
	"github.com/himdhiman/dashboard-backend/services/sentinel-service/models"
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

type GetProductsResponse struct {
	Data  []models.Product `json:"data"`
	Total int              `json:"total"`
	Page  int              `json:"page"`
	Limit int              `json:"limit"`
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
	productsPtr, total, err := uc.Service.GetProducts(ctx, skuCode, pageNumber, limit)
	if err != nil {
		uc.Logger.Error("Error fetching products", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch products"})
		return
	}

	products := make([]models.Product, len(productsPtr))
	for i, p := range productsPtr {
		products[i] = *p
	}

	response := GetProductsResponse{
		Data:  products,
		Total: int(total),
		Page:  pageNumber,
		Limit: limit,
	}

	c.JSON(http.StatusOK, response)
}

// FetchProducts fetches products from the Unicommerce API, runs the task in the background and returns the task ID
func (uc *UnicommerceController) FetchProducts(c *gin.Context) {
	fetchProductsTask := func(params map[string]interface{}) (interface{}, error) {
		ctx := context.Background()
		err := uc.Service.FetchProducts(ctx)
		if err != nil {
			uc.Logger.Error("Error fetching products", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch products"})
			return nil, err
		}
		return nil, nil
	}

	taskID, err := uc.TaskManager.RunTask("FetchProducts", nil, fetchProductsTask)
	if err != nil {
		uc.Logger.Error("Error running fetch products task", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to run fetch products task"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"task_id": taskID})
}

// CreateExportJob creates an export job in the Unicommerce API, runs the task in the background and returns the task ID
func (uc *UnicommerceController) CreateExportJob(c *gin.Context) {
	ctx := context.Background()
	var jobCode string
	cacheErr := uc.Service.TokenManager.Cache.Get(ctx, constants.GetUnicomExportJobCode(), &jobCode)
	if cacheErr == nil && jobCode != "" {
		c.JSON(http.StatusOK, gin.H{"message": "A job is already running"})
		return
	}

	job, err := uc.Service.CreateExportJob(ctx)
	if err != nil {
		uc.Logger.Error("Error creating export job", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create export job"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"job_code": job.JobCode})
}

func (uc *UnicommerceController) SearchProduct(c *gin.Context) {
	ctx := c.Request.Context()

	skuCode := c.Query("skuCode")
	name := c.Query("name")

	products, err := uc.Service.SearchProduct(ctx, skuCode, name)
	if err != nil {
		uc.Logger.Error("Error searching products", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search products"})
		return
	}

	// Create a response with only name and sku
	response := make([]map[string]string, len(products))
	for i, product := range products {
		response[i] = map[string]string{
			"name": product.Name,
			"sku":  product.SKUCode,
		}
	}

	c.JSON(http.StatusOK, gin.H{"products": response})
}
