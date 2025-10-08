package services

import (
	"time"

	products_constants "github.com/himdhiman/dashboard-backend/services/products-service/constants"
	"github.com/himdhiman/dashboard-backend/services/products-service/models"
)

type ExportJobPayloadFactoryFunc func() *models.ExportJobPayload

var ExportJobPayloadFactory = map[string]ExportJobPayloadFactoryFunc{
	products_constants.PRODUCTS_EXPORT_JOB_CODE: func() *models.ExportJobPayload {
		return &models.ExportJobPayload{
			ExportJobTypeName: "Item Master",
			ExportColumns:     []string{"skuCode", "itemName", "imageUrl", "type", "skuType", "itemType_Primary_Vendor"},
			ExportFilters:     nil,
			Frequency:         "ONETIME",
			ReportName:        time.Now().Format("2006-01-02 15:04:05") + "_" + products_constants.PRODUCTS_EXPORT_JOB_CODE,
		}
	},
	products_constants.BUNDLES_EXPORT_JOB_CODE: func() *models.ExportJobPayload {
		return &models.ExportJobPayload{
			ExportJobTypeName: "Item Master",
			ExportColumns:     []string{"skuCode", "itemType_UDF9"},
			ExportFilters:     nil,
			Frequency:         "ONETIME",
			ReportName:        time.Now().Format("2006-01-02 15:04:05") + "_" + products_constants.BUNDLES_EXPORT_JOB_CODE,
		}
	},
	// Add more job codes and their payload generators here
}
