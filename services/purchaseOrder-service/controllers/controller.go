package controllers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	constants "github.com/himdhiman/dashboard-backend/libs/constants"
	"github.com/himdhiman/dashboard-backend/libs/logger"
	"github.com/himdhiman/dashboard-backend/services/purchaseOrder-service/dto"
	"github.com/himdhiman/dashboard-backend/services/purchaseOrder-service/mappers"
	"github.com/himdhiman/dashboard-backend/services/purchaseOrder-service/models"
	"github.com/himdhiman/dashboard-backend/services/purchaseOrder-service/services"
	"github.com/mitchellh/mapstructure"
)

type PurchaseOrderController struct {
	Logger  logger.ILogger
	Service *services.PurchaseOrderService
}

func NewPurchaseOrderController(logger logger.ILogger, service *services.PurchaseOrderService) *PurchaseOrderController {
	return &PurchaseOrderController{
		Logger:  logger,
		Service: service,
	}
}

func (poc *PurchaseOrderController) CreatePurchaseOrder(c *gin.Context) {
	ctx := c.Request.Context()
	correlationID := c.GetHeader(string(constants.CorrelationID))
	if correlationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": constants.ErrMissingCorrelationID})
		return
	}
	ctx = context.WithValue(ctx, constants.CorrelationID, correlationID)

	poc.Logger.Info("Creating purchase order", "correlationID", correlationID)

	var dto dto.CreatePurchaseOrderDTO

	if err := c.ShouldBindJSON(&dto); err != nil {
		poc.Logger.Error("Error binding JSON", "error", err, "correlationID", correlationID)
		c.JSON(http.StatusBadRequest, gin.H{"error": constants.ErrInvalidRequest, "details": err.Error()})
		return
	}

	var purchaseOrder models.PurchaseOrder
	config := &mapstructure.DecoderConfig{
		DecodeHook: mappers.DecodeTimeHookFunc(),
		Result:     &purchaseOrder,
	}
	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		poc.Logger.Error("Error creating decoder", "error", err, "correlationID", correlationID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": constants.ErrInternalServer})
		return
	}

	if err := decoder.Decode(dto); err != nil {
		poc.Logger.Error("Error mapping DTO to model", "error", err, "correlationID", correlationID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": constants.ErrInternalServer})
		return
	}

	poc.Logger.Info("Creating purchase order in service", "correlationID", correlationID)
	err = poc.Service.CreatePurchaseOrder(ctx, &purchaseOrder)
	if err != nil {
		poc.Logger.Error("Error creating purchase order", "error", err, "correlationID", correlationID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": constants.ErrInternalServer})
		return
	}

	poc.Logger.Info("Purchase order created successfully", "orderNumber", purchaseOrder.PONumber, "correlationID", correlationID)
	c.JSON(http.StatusCreated, gin.H{"message": "Purchase order created successfully", "orderNumber": purchaseOrder.PONumber})
}

func (poc *PurchaseOrderController) UpdatePurchaseOrder(c *gin.Context) {
	ctx := c.Request.Context()
	correlationID := c.GetHeader(string(constants.CorrelationID))
	if correlationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": constants.ErrMissingCorrelationID})
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
		poc.Logger.Error("Error binding JSON", "error", err, "correlationID", correlationID)
		c.JSON(http.StatusBadRequest, gin.H{"error": constants.ErrInvalidRequest, "details": err.Error()})
		return
	}

	err := poc.Service.UpdatePurchaseOrder(ctx, poNumber, updates)
	if err != nil {
		poc.Logger.Error("Error updating purchase order", "error", err, "correlationID", correlationID)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to update purchase order" + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Purchase order updated successfully"})
}

func (poc *PurchaseOrderController) GetPurchaseOrders(c *gin.Context) {
	ctx := c.Request.Context()
	correlationID := c.GetHeader(string(constants.CorrelationID))
	if correlationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": constants.ErrMissingCorrelationID})
		return
	}
	ctx = context.WithValue(ctx, constants.CorrelationID, correlationID)

	// Parse query parameters
	pageNumberStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")
	orderNumber := c.DefaultQuery("orderNumber", "")

	pageNumber, err := strconv.Atoi(pageNumberStr)
	if err != nil || pageNumber < 1 {
		poc.Logger.Warn("Invalid page number, defaulting to 1", "pageNumber", pageNumberStr, "correlationID", correlationID)
		pageNumber = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		poc.Logger.Warn("Invalid limit, defaulting to 10", "limit", limitStr, "correlationID", correlationID)
		limit = 10
	}

	// Fetch purchase orders
	purchaseOrdersPtr, total, err := poc.Service.GetPurchaseOrders(ctx, orderNumber, pageNumber, limit)
	if err != nil {
		poc.Logger.Error("Error fetching purchase orders", "error", err, "correlationID", correlationID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": constants.ErrInternalServer})
		return
	}

	if len(purchaseOrdersPtr) == 0 {
		poc.Logger.Info("No purchase orders found", "orderNumber", orderNumber, "correlationID", correlationID)
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

	poc.Logger.Info("Successfully fetched purchase orders", "total", total, "page", pageNumber, "limit", limit, "correlationID", correlationID)
	c.JSON(http.StatusOK, response)
}

func (poc *PurchaseOrderController) DeletePurchaseOrder(c *gin.Context) {
	ctx := c.Request.Context()
	correlationID := c.GetHeader(string(constants.CorrelationID))
	if correlationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": constants.ErrMissingCorrelationID})
		return
	}
	ctx = context.WithValue(ctx, constants.CorrelationID, correlationID)

	poNumber := c.Query("poNumber")
	if poNumber == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing purchase order number"})
		return
	}

	err := poc.Service.DeletePurchaseOrder(ctx, poNumber)
	if err != nil {
		poc.Logger.Error("Error deleting purchase order", "error", err, "correlationID", correlationID)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to delete purchase order" + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Purchase order deleted successfully"})
}

func (poc *PurchaseOrderController) DeletePurchaseOrderProduct(c *gin.Context) {
	ctx := c.Request.Context()
	correlationID := c.GetHeader(string(constants.CorrelationID))
	if correlationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": constants.ErrMissingCorrelationID})
		return
	}
	ctx = context.WithValue(ctx, constants.CorrelationID, correlationID)

	poNumber := c.Query("poNumber")
	if poNumber == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing purchase order number"})
		return
	}

	skuCode := c.Query("skuCode")
	if skuCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing SKU code"})
		return
	}

	err := poc.Service.DeleteProductFromPurchaseOrder(ctx, poNumber, skuCode)
	if err != nil {
		poc.Logger.Error("Error deleting product from purchase order", "error", err, "correlationID", correlationID)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to delete product from purchase order" + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product deleted from purchase order successfully"})
}
