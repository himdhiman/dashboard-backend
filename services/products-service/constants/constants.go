package constants

const (
	SERVICE_CODE = "PRODUCTS_SERVICE"

	MANAGEMENT_SERVICE_CODE = "PRODUCTS_MANAGEMENT_SERVICE"

	UNICOM_API_CODE = "UNICOM_SALTY"

	UnicommerceProductsCollection = "UnicommerceProducts"
	ProductsSchedulersCollection  = "ProductSchedulers"
	ProductsBundlesCollection     = "ProductBundles"
	ShelfwiseInventoryCollection  = "ShelfwiseInventory"

	// Job Codes

	PRODUCTS_EXPORT_JOB_CODE            = "products_export_job"
	BUNDLES_EXPORT_JOB_CODE             = "bundles_export_job"
	SHELFWISE_INVENTORY_EXPORT_JOB_CODE = "shelfwise_inventory_export_job"
	// Add more job codes here

	// API Codes
	API_CODE_UNICOM_FETCH_PRODUCTS    = "FETCH_PRODUCTS"
	API_CODE_UNICOM_CREATE_JOB        = "CREATE_EXPORT_JOB"
	API_CODE_UNICOM_EXPORT_JOB_STATUS = "EXPORT_JOB_STATUS"
	API_CODE_GET_INVENTORY_SNAPSHOT   = "GET_INVENTORY_SNAPSHOT"
	API_CODE_ADJUST_INVENTORY         = "ADJUST_INVENTORY"
)

// ValidExportJobCodes is a set of allowed export job codes
var ValidExportJobCodes = map[string]struct{}{
	PRODUCTS_EXPORT_JOB_CODE:            {},
	BUNDLES_EXPORT_JOB_CODE:             {},
	SHELFWISE_INVENTORY_EXPORT_JOB_CODE: {},
	// add more codes here
}
