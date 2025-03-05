package services

type UnicommerceService struct {
	ServiceCode             string
	Logger                  logger.ILogger
	Cache                   cache.Cacher
	ApiClient               *conflux_client.ConfluxAPIClient
	GoogleSheetService      *GoogleSheetsService
	ProductsRepository      *repository.Repository[models.Product]
	PurchaseOrderRepository *repository.Repository[models.PurchaseOrder]
}