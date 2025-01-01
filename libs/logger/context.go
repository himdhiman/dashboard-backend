package logger

// ContextManager handles correlation ID management
type ContextManager struct {
	serviceName string
}

// NewContextManager creates a new context manager
func NewContextManager(serviceName string) *ContextManager {
	if serviceName == "" {
		serviceName = "default_service"
	}
	return &ContextManager{
		serviceName: serviceName,
	}
}

// GetServiceName retrieves the correlation ID from context
func (cm *ContextManager) GetServiceName() string {
	return cm.serviceName
}
