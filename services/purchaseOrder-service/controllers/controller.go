package controllers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
	constants "github.com/himdhiman/dashboard-backend/libs/constants"
	"github.com/himdhiman/dashboard-backend/libs/logger"
	"github.com/himdhiman/dashboard-backend/services/purchaseOrder-service/dto"
	"github.com/himdhiman/dashboard-backend/services/purchaseOrder-service/services"
)

const SuccessPurchaseOrderCreated = "Purchase order created successfully"

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
	correlationID := poc.getCorrelationID(c)

	poc.Logger.Info("Creating purchase order", "correlationID", correlationID)

	var dto dto.CreatePurchaseOrderDTO

	if err := c.ShouldBindJSON(&dto); err != nil {
		poc.Logger.Error("Error binding JSON", "error", err, "correlationID", correlationID)
		poc.respondWithError(c, http.StatusBadRequest, constants.ErrInvalidRequest, err.Error())
		return
	}

	// Validate the DTO
	if err := validator.New().Struct(&dto); err != nil {
		poc.Logger.Error("Validation failed for DTO", "error", err, "correlationID", correlationID)
		poc.respondWithError(c, http.StatusBadRequest, constants.ErrInvalidRequest, err.Error())
		return
	}

	poc.Logger.Info("Creating purchase order in service", "correlationID", correlationID)
	purchaseOrder, err := poc.Service.CreatePurchaseOrder(ctx, &dto)
	if err != nil {
		if err.Error() == constants.ErrInvalidVendor {
			poc.Logger.Error("Error creating purchase order", "error", err, "correlationID", correlationID)
			poc.respondWithError(c, http.StatusBadRequest, constants.ErrInvalidRequest, err.Error())
			return
		}
		poc.Logger.Error("Error creating purchase order", "error", err, "correlationID", correlationID)
		poc.respondWithError(c, http.StatusBadRequest, constants.ErrInvalidRequest, err.Error())
		return
	}

	poc.Logger.Info("Purchase order created successfully", "orderNumber", purchaseOrder.ID, "correlationID", correlationID)
	poc.respondWithSuccess(c, http.StatusCreated, SuccessPurchaseOrderCreated, purchaseOrder)

}

func (poc *PurchaseOrderController) UpdatePurchaseOrder(c *gin.Context) {
	correlationID := poc.getCorrelationID(c)
	ctx := c.Request.Context()

	poNumber := c.DefaultQuery("poID", "")
	if poNumber == "" {
		poc.Logger.Error("Missing purchase order number", "correlationID", correlationID)
		poc.respondWithError(c, http.StatusBadRequest, "Missing purchase order number", nil)
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		poc.Logger.Error("Error binding JSON", "error", err, "correlationID", correlationID)
		poc.respondWithError(c, http.StatusBadRequest, constants.ErrInvalidRequest, err.Error())
		return
	}

	err := poc.Service.UpdatePurchaseOrder(ctx, poNumber, updates)
	if err != nil {
		poc.Logger.Error("Error updating purchase order", "error", err, "correlationID", correlationID)
		poc.respondWithError(c, http.StatusBadRequest, "Failed to update purchase order", err.Error())
		return
	}

	poc.Logger.Info("Purchase order updated successfully", "correlationID", correlationID)
	poc.respondWithSuccess(c, http.StatusOK, "Purchase order updated successfully", nil)
}

func (poc *PurchaseOrderController) ListPurchaseOrders(c *gin.Context) {
	correlationID := poc.getCorrelationID(c)
	ctx := c.Request.Context()

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
	purchaseOrders, total, err := poc.Service.ListPurchaseOrders(ctx, orderNumber, pageNumber, limit)
	if err != nil {
		poc.Logger.Error("Error fetching purchase orders", "error", err, "correlationID", correlationID)
		poc.respondWithError(c, http.StatusBadRequest, "Failed to fetch purchase orders", err.Error())
		return
	}

	if len(purchaseOrders) == 0 {
		poc.Logger.Info("No purchase orders found", "orderNumber", orderNumber, "correlationID", correlationID)
		poc.respondWithError(c, http.StatusNotFound, "No purchase orders found", nil)
		return
	}

	response := struct {
		PurchaseOrders []dto.ListPurchaseOrdersDTO `json:"purchaseOrders"`
		Total          int                         `json:"total"`
		Page           int                         `json:"page"`
		Limit          int                         `json:"limit"`
	}{
		PurchaseOrders: purchaseOrders,
		Total:          int(total),
		Page:           pageNumber,
		Limit:          limit,
	}

	poc.respondWithSuccess(c, http.StatusOK, "Purchase Order List fetched Successfully", response)
}

func (poc *PurchaseOrderController) GetPurchaseOrder(c *gin.Context) {
	correlationID := poc.getCorrelationID(c)
	ctx := c.Request.Context()

	poID := c.DefaultQuery("poID", "")
	if poID == "" {
		poc.Logger.Error("Missing purchase order id", "correlationID", correlationID)
		poc.respondWithError(c, http.StatusBadRequest, "Missing purchase order id", nil)
		return
	}

	purchaseOrder, err := poc.Service.GetPurchaseOrder(ctx, poID)
	if err != nil {
		poc.Logger.Error("Error fetching purchase order", "error", err, "correlationID", correlationID)
		poc.respondWithError(c, http.StatusBadRequest, "Failed to fetch purchase order", err.Error())
		return
	}

	poc.respondWithSuccess(c, http.StatusOK, "Purchase order fetched successfully", purchaseOrder)
}

func (poc *PurchaseOrderController) DeletePurchaseOrder(c *gin.Context) {
	correlationID := poc.getCorrelationID(c)
	ctx := c.Request.Context()

	poID := c.DefaultQuery("poID", "")
	if poID == "" {
		poc.Logger.Error("Missing purchase order id", "correlationID", correlationID)
		poc.respondWithError(c, http.StatusBadRequest, "Missing purchase order id", nil)
		return
	}

	err := poc.Service.DeletePurchaseOrder(ctx, poID)
	if err != nil {
		poc.Logger.Error("Error deleting purchase order", "error", err, "correlationID", correlationID)
		poc.respondWithError(c, http.StatusBadRequest, "Failed to delete purchase order", err.Error())
		return
	}

	poc.Logger.Info("Purchase order deleted successfully", "correlationID", correlationID)
	poc.respondWithSuccess(c, http.StatusOK, "Purchase order deleted successfully", nil)
}

func (poc *PurchaseOrderController) CreatePurchaseOrderProduct(c *gin.Context) {
	correlationID := poc.getCorrelationID(c)
	ctx := c.Request.Context()

	poc.Logger.Info("Creating purchase order product", "correlationID", correlationID)

	var dto dto.CreatePurchaseOrderProductDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		poc.Logger.Error("Error binding JSON", "error", err, "correlationID", correlationID)
		poc.respondWithError(c, http.StatusBadRequest, constants.ErrInvalidRequest, err.Error())
		return
	}

	poc.Logger.Info("Validating DTO", "correlationID", correlationID)
	if err := validator.New().Struct(&dto); err != nil {
		poc.Logger.Error("Validation failed for DTO", "error", err, "correlationID ", correlationID)
		poc.respondWithError(c, http.StatusBadRequest, constants.ErrInvalidRequest, err.Error())
		return
	}

	poID := c.DefaultQuery("poID", "")
	if poID == "" {
		poc.Logger.Error("Missing purchase order id", "correlationID", correlationID)
		poc.respondWithError(c, http.StatusBadRequest, "Missing purchase order ID", nil)
		return
	}

	poc.Logger.Info("Creating purchase order product in service", "correlationID", correlationID)
	purchaseOrderProduct, err := poc.Service.AddProductToPurchaseOrder(ctx, poID, &dto)
	if err != nil {
		poc.Logger.Error("Error creating purchase order product", "error", err, "correlationID", correlationID)
		poc.respondWithError(c, http.StatusBadRequest, "Failed to create purchase order product", err.Error())
		return
	}

	poc.Logger.Info("Purchase order product created successfully", "correlationID", correlationID)
	c.JSON(http.StatusCreated, purchaseOrderProduct)
}

func (poc *PurchaseOrderController) UpdatePurchaseOrderProduct(c *gin.Context) {
	correlationID := poc.getCorrelationID(c)
	ctx := c.Request.Context()

	poc.Logger.Info("Updating purchase order product", "correlationID", correlationID)

	poID := c.DefaultQuery("poID", "")
	if poID == "" {
		poc.Logger.Error("Missing purchase order ID", "correlationID", correlationID)
		poc.respondWithError(c, http.StatusBadRequest, "Missing purchase order ID", nil)
		return
	}

	productID := c.DefaultQuery("productID", "")
	if productID == "" {
		poc.Logger.Error("Missing product ID", "correlationID", correlationID)
		poc.respondWithError(c, http.StatusBadRequest, "Missing product ID", nil)
		return
	}

	var updates map[string]interface{}
	err := c.BindJSON(&updates)
	if err != nil {
		poc.Logger.Error("Error binding JSON", "error", err, "correlationID", correlationID)
		poc.respondWithError(c, http.StatusBadRequest, constants.ErrInvalidRequest, err.Error())
		return
	}

	err = poc.Service.UpdatePurchaseOrderProduct(ctx, poID, productID, updates)
	if err != nil {
		poc.Logger.Error("Error updating purchase order product", "error", err, "correlationID", correlationID)
		poc.respondWithError(c, http.StatusBadRequest, "Failed to update purchase order product", err.Error())
		return
	}

	poc.Logger.Info("Purchase order product updated successfully", "correlationID", correlationID)
	poc.respondWithSuccess(c, http.StatusOK, "Purchase order product updated successfully", nil)
}

func (poc *PurchaseOrderController) DeletePurchaseOrderProduct(c *gin.Context) {
	correlationID := poc.getCorrelationID(c)
	ctx := c.Request.Context()

	poc.Logger.Info("Deleting purchase order product", "correlationID", correlationID)

	productID := c.DefaultQuery("productID", "")
	if productID == "" {
		poc.Logger.Error("Missing product ID", "correlationID", correlationID)
		poc.respondWithError(c, http.StatusBadRequest, "Missing product ID", nil)
		return
	}

	err := poc.Service.DeleteProductFromPurchaseOrder(ctx, productID)
	if err != nil {
		poc.Logger.Error("Error deleting product from purchase order", "error", err, "correlationID", correlationID)
		poc.respondWithError(c, http.StatusBadRequest, "Failed to delete product from purchase order", err.Error())
		return
	}

	poc.respondWithSuccess(c, http.StatusOK, "Product deleted from purchase order successfully", nil)
}

// Helper function to get the correlation ID from the request

func (poc *PurchaseOrderController) getCorrelationID(c *gin.Context) string {
	return strings.TrimSpace(c.GetHeader(string(constants.CorrelationID)))
}

// Helper function to respond with an error
func (poc *PurchaseOrderController) respondWithError(c *gin.Context, statusCode int, message string, data interface{}) {
	resp := &constants.APIResponse{
		Success: false,
		Message: message,
		Data:    data,
	}
	c.JSON(statusCode, resp)
}

// Helper function to respond with success
func (poc *PurchaseOrderController) respondWithSuccess(c *gin.Context, statusCode int, message string, data interface{}) {
	resp := &constants.APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	}
	c.JSON(statusCode, resp)
}
