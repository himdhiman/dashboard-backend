package constants

type ContextKey string

// HTTP Header constants
const (
	// CorrelationID is the header key for tracking request correlation ID
	CorrelationID ContextKey = "X-Correlation-ID"

	// ContentType is the header key for content type
	ContentType = "Content-Type"

	// Authorization is the header key for authorization
	Authorization = "Authorization"

	// Bearer is the authorization type for bearer tokens
	Bearer = "Bearer"

	// Basic is the authorization type for basic auth
	Basic = "Basic"
)

// Error messages
const (
	ErrMissingCorrelationID = "missing correlation ID"
	ErrInvalidRequest       = "invalid request"
	ErrInternalServer       = "internal server error"
	ErrUnauthorized         = "unauthorized"
	ErrForbidden            = "forbidden"
	ErrNotFound             = "not found"
	ErrInvalidVendor        = "invalid vendor"
)

// Collection Names
const (
	ConfluxApisCollection = "ConfluxApis"
	UnicommerceProductsCollection = "UnicommerceProducts"
	ProductsSchedulersCollection = "ProductSchedulers"
	TaskCollection = "DashboardTasks"
	PurchaseOrderCollection = "PurchaseOrders"

)
