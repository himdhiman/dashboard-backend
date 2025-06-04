package controllers

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/himdhiman/dashboard-backend/libs/constants"
	"github.com/himdhiman/dashboard-backend/libs/logger"
	products_constants "github.com/himdhiman/dashboard-backend/services/products-service/constants"
	"github.com/himdhiman/dashboard-backend/services/products-service/dto"
	"github.com/himdhiman/dashboard-backend/services/products-service/services"
)

type UnicommerceController struct {
	Logger  logger.ILogger
	Service *services.UnicommerceProductsService
}

func NewUnicommerceController(logger logger.ILogger, service *services.UnicommerceProductsService) *UnicommerceController {
	return &UnicommerceController{
		Logger:  logger,
		Service: service,
	}
}

// CreateExportJob creates an export job in the Unicommerce API, runs the task in the background and returns the task ID
func (uc *UnicommerceController) CreateExportJob(c *gin.Context) {
	ctx := c.Request.Context()
	correlationID := c.GetHeader(string(constants.CorrelationID))
	if correlationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": constants.ErrMissingCorrelationID})
		return
	}
	ctx = context.WithValue(ctx, constants.CorrelationID, correlationID)
	jobCode, cacheErr := uc.Service.FetchFromCache(ctx, products_constants.EXPORT_JOB_CODE, "")
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

// AdjustUnicommerceInventory handles the adjustment of Unicommerce inventory based on the provided data.
func (uc *UnicommerceController) AdjustUnicommerceInventory(c *gin.Context) {
	correlationID := uc.getCorrelationID(c)
	ctx := c.Request.Context()

	uc.Logger.Info("Received request to adjust Unicommerce inventory", "correlationID", correlationID)

	var data dto.ProductPayloadDTO
	if err := c.ShouldBindJSON(&data); err != nil {
		uc.Logger.Error("Error binding JSON", "error", err)
		uc.respondWithError(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	err := uc.Service.AdjustUnicommerceInventory(ctx, data)
	if err != nil {
		uc.Logger.Error("Error adjusting Unicommerce inventory", "error", err)
		uc.respondWithError(c, http.StatusBadRequest, "Failed to adjust inventory", err.Error())
		return
	}

	uc.Logger.Info("Successfully adjusted Unicommerce inventory", "correlationID", correlationID)
	uc.respondWithSuccess(c, http.StatusOK, "Inventory adjustment successful", nil)
}

func (uc *UnicommerceController) getCorrelationID(c *gin.Context) string {
	return strings.TrimSpace(c.GetHeader(string(constants.CorrelationID)))
}

// Helper function to respond with an error
func (uc *UnicommerceController) respondWithError(c *gin.Context, statusCode int, message string, data interface{}) {
	resp := &constants.APIResponse{
		Success: false,
		Message: message,
		Data:    data,
	}
	c.JSON(statusCode, resp)
}

// Helper function to respond with success
func (uc *UnicommerceController) respondWithSuccess(c *gin.Context, statusCode int, message string, data interface{}) {
	resp := &constants.APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	}
	c.JSON(statusCode, resp)
}
