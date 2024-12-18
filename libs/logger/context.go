package logger

import (
	"context"

	"github.com/google/uuid"
)

// ContextManager handles correlation ID management
type ContextManager struct {
	correlationKey string
}

// NewContextManager creates a new context manager
func NewContextManager(key string) *ContextManager {
	if key == "" {
		key = "correlation_id"
	}
	return &ContextManager{
		correlationKey: key,
	}
}

// ExtractOrCreateCorrelationID extracts existing correlation ID or creates a new one
func (cm *ContextManager) ExtractOrCreateCorrelationID(ctx context.Context) (context.Context, string) {
	// Check if correlation ID exists
	if correlationID, ok := ctx.Value(cm.correlationKey).(string); ok && correlationID != "" {
		return ctx, correlationID
	}

	// Generate new correlation ID
	newCorrelationID := uuid.New().String()

	// Create new context with correlation ID
	newCtx := context.WithValue(ctx, cm.correlationKey, newCorrelationID)

	return newCtx, newCorrelationID
}

// GetCorrelationID retrieves the correlation ID from context
func (cm *ContextManager) GetCorrelationID(ctx context.Context) string {
	if correlationID, ok := ctx.Value(cm.correlationKey).(string); ok {
		return correlationID
	}
	return ""
}
