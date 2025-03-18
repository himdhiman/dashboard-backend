package services

import (
	"context"

	"github.com/himdhiman/dashboard-backend/libs/logger"
	mongo_models "github.com/himdhiman/dashboard-backend/libs/mongo/models"
	"github.com/himdhiman/dashboard-backend/libs/mongo/repository"
	products_constants "github.com/himdhiman/dashboard-backend/services/products-service/constants"
	"github.com/himdhiman/dashboard-backend/services/products-service/models"
	"github.com/himdhiman/dashboard-backend/services/products-service/utils"
)

type ManagementService struct {
	ServiceCode        string
	Logger             logger.ILogger
	ProductsRepository *repository.Repository[models.Product]
}

func NewManagementService(logger logger.ILogger, productsCollection *mongo_models.MongoCollection) *ManagementService {

	productsRepo := repository.Repository[models.Product]{Collection: productsCollection}

	return &ManagementService{
		ServiceCode:        products_constants.MANAGEMENT_SERVICE_CODE,
		Logger:             logger,
		ProductsRepository: &productsRepo,
	}
}

func (s *ManagementService) UpdateLastProcuredRMBPriceFromCSV(ctx context.Context, filePath string) error {
	csvRecords, err := utils.ReadCSV(filePath)
	if err != nil {
		s.Logger.Error("Error reading CSV file", "error", err)
		return err
	}

	for _, record := range csvRecords {
		filter := map[string]interface{}{
			"skuCode":       record.SKU,
			"primaryVendor": record.Vendor,
		}
		update := map[string]interface{}{
			"lastProcuredRmbPrice": record.Price,
		}

		_, err := s.ProductsRepository.Update(ctx, filter, update)
		if err != nil {
			s.Logger.Error("Error updating product", "SKU", record.SKU, "Vendor", record.Vendor, "error", err)
			return err
		}
		s.Logger.Info("Updated product", "SKU", record.SKU, "Vendor", record.Vendor, "lastProcuredRMBPrice", record.Price)
	}

	return nil
}
