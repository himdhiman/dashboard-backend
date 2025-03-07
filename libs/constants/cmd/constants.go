package constants

type ContextKey string

// HTTP Header constants
const (
	// CorrelationID is the header key for tracking request correlation ID
	CorrelationID ContextKey = "X-Correlation-ID"

	// ContentType is the header key for content type
	ContentType = "Content-Type"

	// ApplicationJSON is the content type for JSON responses
	ApplicationJSON = "application/json"

	// Authorization is the header key for authorization token
	Authorization = "Authorization"
)

// Request parameter constants
const (
	// DefaultPageSize is the default number of items per page
	DefaultPageSize = 10

	// MaxPageSize is the maximum number of items allowed per page
	MaxPageSize = 100
)

// Status message constants
const (
	// StatusSuccess indicates successful operation
	StatusSuccess = "SUCCESS"

	// StatusFailed indicates a failed operation
	StatusFailed = "FAILED"

	// StatusPending indicates a pending operation
	StatusPending = "PENDING"
)

// Error message constants
const (
	// ErrInvalidRequest indicates invalid request parameters
	ErrInvalidRequest = "Invalid request parameters"

	// ErrInternalServer indicates an internal server error
	ErrInternalServer = "Internal server error"

	// ErrResourceNotFound indicates requested resource was not found
	ErrResourceNotFound = "Resource not found"

	// ErrUnauthorized indicates unauthorized access attempt
	ErrUnauthorized = "Unauthorized access"
)
