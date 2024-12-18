package mongo

// PaginationOptions represents pagination settings
type PaginationOptions struct {
	Page     int64
	PageSize int64
}

// FilterOptions represents a basic filter
type FilterOptions map[string]interface{}
