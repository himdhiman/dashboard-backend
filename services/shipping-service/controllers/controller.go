package controllers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/himdhiman/dashboard-backend/libs/constants"
	"github.com/himdhiman/dashboard-backend/libs/logger"

	"github.com/himdhiman/dashboard-backend/services/shipping-service/dto"
	"github.com/himdhiman/dashboard-backend/services/shipping-service/services"
)

type ShippingController struct {
	Logger  logger.ILogger
	Service *services.ShippingService
}

func NewShippingController(logger logger.ILogger, service *services.ShippingService) *ShippingController {
	return &ShippingController{
		Logger:  logger,
		Service: service,
	}
}

func (sc *ShippingController) ListShippingMarks(c *gin.Context) {
	correlationID := sc.getCorrelationID(c)
	ctx := c.Request.Context()

	// Parse query parameters
	pageNumberStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")
	shippingMark := c.DefaultQuery("shippingMark", "")

	pageNumber, err := strconv.Atoi(pageNumberStr)
	if err != nil || pageNumber < 1 {
		sc.Logger.Warn("Invalid page number, defaulting to 1", "pageNumber", pageNumberStr, "correlationID", correlationID)
		pageNumber = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		sc.Logger.Warn("Invalid limit, defaulting to 10", "limit", limitStr, "correlationID", correlationID)
		limit = 10
	}

	// Fetch Shipping Marks
	shippingMarks, total, err := sc.Service.ListShippingMarks(ctx, shippingMark, pageNumber, limit)
	if err != nil {
		sc.Logger.Error("Error fetching Shipping Marks", "error", err, "correlationID", correlationID)
		sc.respondWithError(c, http.StatusBadRequest, "Failed to fetch Shipping Marks", err.Error())
		return
	}

	if len(shippingMarks) == 0 {
		sc.Logger.Info("No shipping marks found", "correlationID", correlationID)
		sc.respondWithError(c, http.StatusNotFound, "No shipping marks found", nil)
		return
	}

	response := struct {
		ShippingMarks []dto.ListShippingMarksDTO `json:"shippingMarks"`
		Total         int                        `json:"total"`
		Page          int                        `json:"page"`
		Limit         int                        `json:"limit"`
	}{
		ShippingMarks: shippingMarks,
		Total:         int(total),
		Page:          pageNumber,
		Limit:         limit,
	}

	sc.respondWithSuccess(c, http.StatusOK, "Purchase Order List fetched Successfully", response)
}

func (sc *ShippingController) GetShippingMark(c *gin.Context) {
	correlationID := sc.getCorrelationID(c)
	ctx := c.Request.Context()

	shippingMarkID := c.DefaultQuery("shippingMarkID", "")
	if shippingMarkID == "" {
		sc.Logger.Error("Missing Shipping Mark ID", "correlationID", correlationID)
		sc.respondWithError(c, http.StatusBadRequest, "Missing shipping mark ID", nil)
		return
	}

	shippingMark, err := sc.Service.GetShippingMark(ctx, shippingMarkID)
	if err != nil {
		sc.Logger.Error("Error fetching Shipping Mark", "error", err, "correlationID", correlationID)
		sc.respondWithError(c, http.StatusBadRequest, "Failed to fetch Shipping Mark", err.Error())
		return
	}

	sc.respondWithSuccess(c, http.StatusOK, "Shipping Mark fetched successfully", shippingMark)
}

func (sc *ShippingController) UpdateShippingMark(c *gin.Context) {
	correlationID := sc.getCorrelationID(c)
	ctx := c.Request.Context()

	shippingMarkID := c.DefaultQuery("shippingMarkID", "")
	if shippingMarkID == "" {
		sc.Logger.Error("Missing Shipping Mark ID", "correlationID", correlationID)
		sc.respondWithError(c, http.StatusBadRequest, "Missing shipping mark ID", nil)
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		sc.Logger.Error("Error binding JSON", "error", err, "correlationID", correlationID)
		sc.respondWithError(c, http.StatusBadRequest, constants.ErrInvalidRequest, err.Error())
		return
	}

	err := sc.Service.UpdateShippingMark(ctx, shippingMarkID, updates)
	if err != nil {
		sc.Logger.Error("Error updating Shipping Mark", "error", err, "correlationID", correlationID)
		sc.respondWithError(c, http.StatusBadRequest, "Failed to update Shipping Mark", err.Error())
		return
	}

	sc.respondWithSuccess(c, http.StatusOK, "Shipping Mark updated successfully", nil)
}

func (sc *ShippingController) getCorrelationID(c *gin.Context) string {
	return strings.TrimSpace(c.GetHeader(string(constants.CorrelationID)))
}

// Helper function to respond with an error
func (sc *ShippingController) respondWithError(c *gin.Context, statusCode int, message string, data interface{}) {
	resp := &constants.APIResponse{
		Success: false,
		Message: message,
		Data:    data,
	}
	c.JSON(statusCode, resp)
}

// Helper function to respond with success
func (sc *ShippingController) respondWithSuccess(c *gin.Context, statusCode int, message string, data interface{}) {
	resp := &constants.APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	}
	c.JSON(statusCode, resp)
}
