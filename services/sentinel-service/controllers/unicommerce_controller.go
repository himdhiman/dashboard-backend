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

func (uc *UnicommerceController) CreatePurchaseOrder(c *gin.Context) {
	ctx := c.Request.Context()
	correlationID := c.GetHeader("X-Correlation-ID")
	if correlationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing correlation ID"})
		return
	}
	ctx = context.WithValue(ctx, constants.CorrelationID, correlationID)

	uc.Logger.Info("Creating purchase order", "correlationID", correlationID)

	var dto dto.CreatePurchaseOrderDTO

	if err := c.ShouldBindJSON(&dto); err != nil {
		uc.Logger.Error("Error binding JSON", "error", err, "correlationID", correlationID)
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
		uc.Logger.Error("Error creating decoder", "error", err, "correlationID", correlationID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create decoder"})
		return
	}

	if err := decoder.Decode(dto); err != nil {
		uc.Logger.Error("Error mapping DTO to model", "error", err, "correlationID", correlationID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to map request payload"})
		return
	}

	uc.Logger.Info("Creating purchase order in service", "correlationID", correlationID)
	err = uc.Service.CreatePurchaseOrder(ctx, &purchaseOrder)
	if err != nil {
		uc.Logger.Error("Error creating purchase order", "error", err, "correlationID", correlationID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create purchase order"})
		return
	}

	uc.Logger.Info("Purchase order created successfully", "orderNumber", purchaseOrder.PONumber, "correlationID", correlationID)
	c.JSON(http.StatusCreated, gin.H{"message": "Purchase order created successfully", "orderNumber": purchaseOrder.PONumber})
}

func (uc *UnicommerceController) UpdatePurchaseOrder(c *gin.Context) {
	ctx := c.Request.Context()
	correlationID := c.GetHeader("X-Correlation-ID")
	if correlationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing correlation ID"})
		return
	}
	ctx = context.WithValue(ctx, constants.CorrelationID, correlationID)

	poNumber := c.DefaultQuery("poNumber", "")
	if poNumber == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing purchase order number"})
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		uc.Logger.Error("Error binding JSON", "error", err, "correlationID", correlationID)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload", "details": err.Error()})
		return
	}

	err := uc.Service.UpdatePurchaseOrder(ctx, poNumber, updates)
	if err != nil {
		uc.Logger.Error("Error updating purchase order", "error", err, "correlationID", correlationID)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to update purchase order" + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Purchase order updated successfully"})
}

func (uc *UnicommerceController) GetPurchaseOrders(c *gin.Context) {
	ctx := c.Request.Context()
	correlationID := c.GetHeader("X-Correlation-ID")
	if correlationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing correlation ID"})
		return
	}
	ctx = context.WithValue(ctx, constants.CorrelationID, correlationID)

	// Parse query parameters
	pageNumberStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")
	orderNumber := c.DefaultQuery("orderNumber", "")

	pageNumber, err := strconv.Atoi(pageNumberStr)
	if err != nil || pageNumber < 1 {
		uc.Logger.Warn("Invalid page number, defaulting to 1", "pageNumber", pageNumberStr, "correlationID", correlationID)
		pageNumber = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		uc.Logger.Warn("Invalid limit, defaulting to 10", "limit", limitStr, "correlationID", correlationID)
		limit = 10
	}

	// Fetch purchase orders
	purchaseOrdersPtr, total, err := uc.Service.GetPurchaseOrders(ctx, orderNumber, pageNumber, limit)
	if err != nil {
		uc.Logger.Error("Error fetching purchase orders", "error", err, "correlationID", correlationID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch purchase orders"})
		return
	}

	if len(purchaseOrdersPtr) == 0 {
		uc.Logger.Info("No purchase orders found", "orderNumber", orderNumber, "correlationID", correlationID)
		c.JSON(http.StatusOK, gin.H{"data": []models.PurchaseOrder{}, "total": 0, "page": pageNumber, "limit": limit})
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

	uc.Logger.Info("Successfully fetched purchase orders", "total", total, "page", pageNumber, "limit", limit, "correlationID", correlationID)
	c.JSON(http.StatusOK, response)
}
