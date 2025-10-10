package controllers

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/himdhiman/dashboard-backend/libs/constants"
)

func getCorrelationID(c *gin.Context) string {
	return strings.TrimSpace(c.GetHeader(string(constants.CorrelationID)))
}

// Helper function to respond with an error
func respondWithError(c *gin.Context, statusCode int, message string, data interface{}) {
	resp := &constants.APIResponse{
		Success: false,
		Message: message,
		Data:    data,
	}
	c.JSON(statusCode, resp)
}

// Helper function to respond with success
func respondWithSuccess(c *gin.Context, statusCode int, message string, data interface{}) {
	resp := &constants.APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	}
	c.JSON(statusCode, resp)
}
