package controllers

import (
	"context"
	"net/http"

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

	var req struct {
		ExportJobCode string `json:"exportJobCode" form:"exportJobCode" binding:"required"`
	}
	if err := c.ShouldBind(&req); err != nil || req.ExportJobCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing exportJobCode"})
		return
	}

	// Validate export job code
	if _, ok := products_constants.ValidExportJobCodes[req.ExportJobCode]; !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid exportJobCode"})
		return
	}

	jobCode, cacheErr := uc.Service.FetchFromCache(ctx, req.ExportJobCode, "")
	if cacheErr == nil && jobCode != "" {
		c.JSON(http.StatusOK, gin.H{"message": "A job is already running"})
		return
	}

	job, err := uc.Service.CreateExportJobByCode(ctx, req.ExportJobCode)
	if err != nil {
		uc.Logger.Error("Error creating export job", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create export job"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"job_code": job.JobCode})
}

// AdjustUnicommerceInventory handles the adjustment of Unicommerce inventory based on the provided data.
func (uc *UnicommerceController) AdjustUnicommerceInventory(c *gin.Context) {
	correlationID := getCorrelationID(c)
	ctx := c.Request.Context()

	uc.Logger.Info("Received request to adjust Unicommerce inventory", "correlationID", correlationID)

	var data dto.ProductPayloadDTO
	if err := c.ShouldBindJSON(&data); err != nil {
		uc.Logger.Error("Error binding JSON", "error", err)
		respondWithError(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	uc.Logger.Info("Adjusting Unicommerce inventory", "correlationID", correlationID, "Data Recieved", data)

	err := uc.Service.AdjustUnicommerceInventory(ctx, data)
	if err != nil {
		uc.Logger.Error("Error adjusting Unicommerce inventory", "error", err)
		respondWithError(c, http.StatusBadRequest, "Failed to adjust inventory", err.Error())
		return
	}

	uc.Logger.Info("Successfully adjusted Unicommerce inventory", "correlationID", correlationID)
	respondWithSuccess(c, http.StatusOK, "Inventory adjustment successful", nil)
}

// CheckJobStatus checks if a job is running or finished
func (uc *UnicommerceController) CheckJobStatus(c *gin.Context) {
	ctx := c.Request.Context()
	correlationID := c.GetHeader(string(constants.CorrelationID))
	if correlationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": constants.ErrMissingCorrelationID})
		return
	}
	ctx = context.WithValue(ctx, constants.CorrelationID, correlationID)

	var req struct {
		JobCode string `json:"job_code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		uc.Logger.Error("Error binding JSON for job status check", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	status, err := uc.Service.CheckJobStatusByJobCode(ctx, req.JobCode)
	if err != nil {
		uc.Logger.Error("Error checking job status", "jobCode", req.JobCode, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check job status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": status})
}
