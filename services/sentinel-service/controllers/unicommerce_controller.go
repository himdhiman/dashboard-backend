package controllers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/himdhiman/dashboard-backend/libs/logger"
	"github.com/himdhiman/dashboard-backend/libs/task"
	"github.com/himdhiman/dashboard-backend/services/sentinel-service/constants"
	"github.com/himdhiman/dashboard-backend/services/sentinel-service/dto"
	"github.com/himdhiman/dashboard-backend/services/sentinel-service/mappers"
	"github.com/himdhiman/dashboard-backend/services/sentinel-service/models"
	"github.com/himdhiman/dashboard-backend/services/sentinel-service/services"
	"github.com/mitchellh/mapstructure"
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

	var request struct {
		SKUCode string   `json:"skuCode"`
		Name    string   `json:"name"`
		Fields  []string `json:"fields"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		uc.Logger.Error("Error binding JSON", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	products, err := uc.Service.SearchProduct(ctx, request.SKUCode, request.Name)
	if err != nil {
		uc.Logger.Error("Error searching products", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search products"})
		return
	}

	response := make([]map[string]interface{}, len(products))
	for i, product := range products {
		productMap := make(map[string]interface{})
		for _, field := range request.Fields {
			switch field {
			case "name":
				productMap["name"] = product.Name
			case "sku":
				productMap["sku"] = product.SKUCode
			case "imageUrl":
				productMap["imageUrl"] = product.ImageURL
			case "primaryVendor":
				productMap["primaryVendor"] = product.PrimaryVendor
			case "lastProcuredRmbPrice":
				productMap["lastProcuredRmbPrice"] = product.LastProcuredRmbPrice
			case "createdAt":
				productMap["createdAt"] = product.CreatedAt
			case "updatedAt":
				productMap["updatedAt"] = product.UpdatedAt
			default:
				uc.Logger.Warn("Unknown field requested", "field", field)
			}
		}
		response[i] = productMap
	}

	c.JSON(http.StatusOK, gin.H{"products": response})
}

func (uc *UnicommerceController) CreatePurchaseOrder(c *gin.Context) {
	var dto dto.CreatePurchaseOrderDTO

	if err := c.ShouldBindJSON(&dto); err != nil {
		uc.Logger.Error("Error binding JSON", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload", "details": err.Error()})
		return
	}

	var purchaseOrder models.PurchaseOrder
	config := &mapstructure.DecoderConfig{
		DecodeHook: mappers.DecodeTimeHookFunc(),
		Result:     &purchaseOrder,
	}
	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		uc.Logger.Error("Error creating decoder", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create decoder"})
		return
	}

	if err := decoder.Decode(dto); err != nil {
		uc.Logger.Error("Error mapping DTO to model", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to map request payload"})
		return
	}

	err = uc.Service.CreatePurchaseOrder(c.Request.Context(), &purchaseOrder)
	if err != nil {
		uc.Logger.Error("Error creating purchase order", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create purchase order"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Purchase order created successfully", "orderNumber": purchaseOrder.PONumber})
}


func (uc *UnicommerceController) GetPurchaseOrders(c *gin.Context) {
	// Parse query parameters

	pageNumberStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")
	orderNumber := c.DefaultQuery("orderNumber", "")

	pageNumber, err := strconv.Atoi(pageNumberStr)
	if err != nil || pageNumber < 1 {
		pageNumber = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 10
	}

	// Fetch purchase orders
	ctx := c.Request.Context()
	purchaseOrdersPtr, total, err := uc.Service.GetPurchaseOrders(ctx, orderNumber, pageNumber, limit)
	if err != nil {
		uc.Logger.Error("Error fetching purchase orders", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch purchase orders"})
		return
	}

	purchaseOrders := make([]models.PurchaseOrder, len(purchaseOrdersPtr))
	for i, p := range purchaseOrdersPtr {
		purchaseOrders[i] = *p
	}

	response := struct {
		Data  []models.PurchaseOrder `json:"data"`
		Total int                    `json:"total"`
		Page  int                    `json:"page"`
		Limit int                    `json:"limit"`
	}{
		Data:  purchaseOrders,
		Total: int(total),
		Page:  pageNumber,
		Limit: limit,
	}

	c.JSON(http.StatusOK, response)
}
