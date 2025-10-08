package services

import "github.com/himdhiman/dashboard-backend/services/products-service/constants"

// ExportProcessorFactory maps job codes to their processors
var ExportProcessorFactory = map[string]FileProcessorFunc{
	constants.PRODUCTS_EXPORT_JOB_CODE: ProductsExportProcessor,
	// Add more job code to processor mappings here
}
