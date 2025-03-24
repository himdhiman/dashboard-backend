package middlewares

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/himdhiman/dashboard-backend/libs/constants"
)

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-Correlation-ID")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func CorrelationIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		correlationID := strings.TrimSpace(c.GetHeader(string(constants.CorrelationID)))
		if correlationID == "" {
			c.JSON(http.StatusBadRequest, &constants.APIResponse{
				Success: false,
				Message: constants.ErrMissingCorrelationID,
			})
			c.Abort()
			return
		}

		// Inject the correlation ID into the context
		ctx := context.WithValue(c.Request.Context(), constants.CorrelationID, correlationID)
		c.Request = c.Request.WithContext(ctx)

		// Proceed to the next middleware or handler
		c.Next()
	}
}
