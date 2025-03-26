package services

import (
	"context"
	"time"

	"github.com/himdhiman/dashboard-backend/libs/constants"
	"github.com/himdhiman/dashboard-backend/libs/logger"
	"github.com/himdhiman/dashboard-backend/libs/mappers"
	"github.com/himdhiman/dashboard-backend/libs/mongo"
	"github.com/himdhiman/dashboard-backend/libs/mongo/errors"
	mongo_models "github.com/himdhiman/dashboard-backend/libs/mongo/models"
	"github.com/himdhiman/dashboard-backend/libs/mongo/repository"
	"github.com/himdhiman/dashboard-backend/libs/task"
	product_services "github.com/himdhiman/dashboard-backend/services/products-service/services"
	puchaseOrder_models "github.com/himdhiman/dashboard-backend/services/purchaseOrder-service/models"
	"github.com/himdhiman/dashboard-backend/services/shipping-service/dto"
	"github.com/himdhiman/dashboard-backend/services/shipping-service/models"
	"github.com/himdhiman/dashboard-backend/services/shipping-service/utils"
)

type ShippingService struct {
	Logger                          logger.ILogger
	Mapper                          *mappers.Mapper
	TaskManager                     *task.TaskManager
	ProductsService                 product_services.ProductsService
	ShippingMarkRepository          *repository.Repository[models.ShippingMark]
	ShippingProviderRepository      *repository.Repository[models.ShippingProvider]
	PurchaseOrderProductsRepository *repository.Repository[puchaseOrder_models.PurchaseOrderProducts]
}

func NewShippingService(logger logger.ILogger, mapper *mappers.Mapper,
	productsService product_services.ProductsService,
	taskManager *task.TaskManager, shippingMarkRepository *repository.Repository[models.ShippingMark],
	shippingProviderRepository *repository.Repository[models.ShippingProvider],
	purchaseOrderProductsRepository *repository.Repository[puchaseOrder_models.PurchaseOrderProducts]) *ShippingService {
	return &ShippingService{
		Logger:                          logger,
		Mapper:                          mapper,
		TaskManager:                     taskManager,
		ProductsService:                 productsService,
		ShippingMarkRepository:          shippingMarkRepository,
		ShippingProviderRepository:      shippingProviderRepository,
		PurchaseOrderProductsRepository: purchaseOrderProductsRepository,
	}
}

func (s *ShippingService) CreateShippingMark(ctx context.Context, shippingMark string) (string, error) {
	correlationID := ctx.Value(constants.CorrelationID).(string)
	s.Logger.Info("Creating purchase order", "correlationID", correlationID)

	s.Logger.Info("Creating shipping mark")

	// create chipping mark, if it do not exist
	_, err := s.ShippingMarkRepository.FindOne(ctx, map[string]interface{}{"shippingMark": shippingMark}, nil)
	if err != nil {
		if err == errors.ErrDocumentNotFound {
			// create shipping mark

			shippingProvider, err := s.ShippingProviderRepository.FindOne(ctx, map[string]interface{}{"vendorName": "DHL"}, nil)
			if err != nil {
				s.Logger.Error("Failed to find shipping provider", "correlationID", correlationID, "error", err)
				return "", err
			}

			shippingMark := &models.ShippingMark{
				ID:               mongo.NewObjectID(),
				ShippingMark:     shippingMark,
				ShippingProvider: shippingProvider.ID.Hex(),
				Status:           "InTransit",
			}

			// validate the struct using v10 validator
			validator := s.Mapper.GetValidator()
			err = validator.Struct(shippingMark)
			if err != nil {
				s.Logger.Error("Failed to validate shipping mark", "correlationID", correlationID, "error", err)
				return "", err
			}

			_, err = s.ShippingMarkRepository.Create(ctx, shippingMark)
			if err != nil {
				s.Logger.Error("Failed to create shipping mark", "correlationID", correlationID, "error", err)
				return "", err
			}
		} else {
			s.Logger.Error("Failed to find shipping mark", "correlationID", correlationID, "error", err)
			return "", err
		}
	}
	return shippingMark, nil
}

func (s *ShippingService) ListShippingProviders(ctx context.Context) ([]*models.ShippingProvider, error) {
	correlationID := ctx.Value(constants.CorrelationID).(string)
	s.Logger.Info("Listing shipping providers", "correlationID", correlationID)

	shippingProviders, err := s.ShippingProviderRepository.Find(ctx, nil, nil)
	if err != nil {
		s.Logger.Error("Failed to list shipping providers", "correlationID", correlationID, "error", err)
		return nil, err
	}

	return shippingProviders, nil
}

func (s *ShippingService) ListShippingMarks(ctx context.Context, shippingMark string, pageNumber, fieldsPerPage int) ([]dto.ListShippingMarksDTO, int64, error) {
	correlationID := ctx.Value(constants.CorrelationID).(string)
	s.Logger.Info("Listing shipping marks", "correlationID", correlationID)

	filter := map[string]interface{}{}
	if shippingMark != "" {
		filter["shippingMark"] = shippingMark
	}

	// Ensure pageNumber is at least 1
	if pageNumber < 1 {
		pageNumber = 1
	}

	shippingMarks, err := s.ShippingMarkRepository.Find(ctx, filter, &mongo_models.FindOptions{
		Limit: int64(fieldsPerPage),
		Skip:  int64((pageNumber - 1) * fieldsPerPage),
	})
	if err != nil {
		s.Logger.Error("Failed to list shipping marks", "correlationID", correlationID, "error", err)
		return nil, 0, err
	}

	cnt, err := s.ShippingMarkRepository.Count(ctx, filter)
	if err != nil {
		s.Logger.Error("Error fetching Shipping Marks count", "correlationID", correlationID, "error", err)
		return nil, 0, err
	}

	var shippingMarkListDTO []dto.ListShippingMarksDTO

	for _, shippingMark := range shippingMarks {
		var mappedShippingMarkDTO dto.ListShippingMarksDTO
		err := s.Mapper.DecodeWithCustomHook(shippingMark, &mappedShippingMarkDTO, mappers.DecodeObjectIDHookFunc(), mappers.EncodeTimeToStringHookFunc())
		if err != nil {
			s.Logger.Error("Error decoding purchase order", "correlationID", correlationID, "error", err)
			return nil, 0, err
		}
		mappedShippingMarkDTO.DispatchDate = shippingMark.DispatchDate.Format(time.RFC3339)
		mappedShippingMarkDTO.DeliveryDate = shippingMark.DeliveryDate.Format(time.RFC3339)
		mappedShippingMarkDTO.WarehouseRecievingDate = shippingMark.WarehouseRecievingDate.Format(time.RFC3339)

		shippingProvider, err := s.ShippingProviderRepository.FindOne(ctx, map[string]interface{}{"_id": shippingMark.ShippingProvider}, nil)
		if err != nil {
			s.Logger.Error("Error fetching shipping provider", "correlationID", correlationID, "error", err)
			return nil, 0, err
		}

		mappedShippingMarkDTO.ShippingAgent = shippingProvider.VendorName

		shippingMarkListDTO = append(shippingMarkListDTO, mappedShippingMarkDTO)
	}

	return shippingMarkListDTO, cnt, nil
}

func (s *ShippingService) GetShippingMark(ctx context.Context, shippingMarkID string) (*dto.ShippingMarkDTO, error) {
	correlationID := ctx.Value(constants.CorrelationID).(string)
	s.Logger.Info("Getting shipping mark", "correlationID", correlationID)

	shippingMark, err := s.ShippingMarkRepository.FindOne(ctx, map[string]interface{}{"_id": shippingMarkID}, nil)
	if err != nil {
		s.Logger.Error("Failed to find shipping mark", "correlationID", correlationID, "error", err)
		return nil, err
	}

	var shippingMarkDTO dto.ShippingMarkDTO
	err = s.Mapper.DecodeWithCustomHook(shippingMark, &shippingMarkDTO, mappers.DecodeObjectIDHookFunc(), mappers.EncodeTimeToStringHookFunc())
	if err != nil {
		s.Logger.Error("Error decoding purchase order", "correlationID", correlationID, "error", err)
		return nil, err
	}
	shippingMarkDTO.DispatchDate = shippingMark.DispatchDate.Format(time.RFC3339)
	shippingMarkDTO.DeliveryDate = shippingMark.DeliveryDate.Format(time.RFC3339)
	shippingMarkDTO.WarehouseRecievingDate = shippingMark.WarehouseRecievingDate.Format(time.RFC3339)

	shippingProvider, err := s.ShippingProviderRepository.FindOne(ctx, map[string]interface{}{"_id": shippingMark.ShippingProvider}, nil)
	if err != nil {
		s.Logger.Error("Error fetching shipping provider", "correlationID", correlationID, "error", err)
		return nil, err
	}

	shippingMarkDTO.ShippingAgent = shippingProvider.VendorName

	// Fetch all the products with the same shipping mark
	products, err := s.PurchaseOrderProductsRepository.Find(ctx, map[string]interface{}{"shippingMark": shippingMark.ShippingMark}, nil)
	if err != nil {
		s.Logger.Error("Failed to find products", "correlationID", correlationID, "error", err)
		return nil, err
	}

	var shippingMarkProductDTOs []dto.ShippingMarkProductDTO
	for _, product := range products {
		var mappedshippingMarkProduct dto.ShippingMarkProductDTO
		err = s.Mapper.DecodeWithCustomHook(product, &mappedshippingMarkProduct, mappers.DecodeObjectIDHookFunc(), mappers.EncodeTimeToStringHookFunc())
		if err != nil {
			s.Logger.Error("Error decoding purchase order product", "correlationID", correlationID, "error", err)
			return nil, err
		}

		fetchedProduct, err := s.ProductsService.GetProductByID(ctx, product.ProductID)
		if err != nil {
			s.Logger.Error("Error fetching product", "correlationID", correlationID, "error", err)
			return nil, err
		}

		mappedshippingMarkProduct.SKUCode = fetchedProduct.SKUCode
		mappedshippingMarkProduct.ImageURL = fetchedProduct.ImageURL

		mappedshippingMarkProduct.OrderDate = product.OrderDate.Format(time.RFC3339)

		shippingMarkProductDTOs = append(shippingMarkProductDTOs, mappedshippingMarkProduct)
	}

	shippingMarkDTO.Products = shippingMarkProductDTOs

	return &shippingMarkDTO, nil
}

func (s *ShippingService) UpdateShippingMark(ctx context.Context, shippingMarkID string, updates map[string]interface{}) error {
	correlationID := ctx.Value(constants.CorrelationID).(string)
	s.Logger.Info("Updating shipping mark", "correlationID", correlationID)

	shippingMark, err := s.ShippingMarkRepository.FindOne(ctx, map[string]interface{}{"_id": shippingMarkID}, nil)
	if err != nil {
		s.Logger.Error("Failed to find shipping mark", "correlationID", correlationID, "error", err)
		return err
	}

	// Update the fields
	for fieldPath, value := range updates {
		err := constants.SetField(shippingMark, fieldPath, value, utils.AllowedFields)
		if err != nil {
			s.Logger.Error("Error setting field", "correlationID", correlationID, "fieldPath", fieldPath, "error", err)
			return err
		}
	}

	validator := s.Mapper.GetValidator()
	err = validator.Struct(shippingMark)
	if err != nil {
		s.Logger.Error("Error validating purchase order", "correlationID", correlationID, "error", err)
		return err
	}

	if _, ok := updates["ShippingProvider"]; ok {
		_, err := s.ShippingProviderRepository.FindOne(ctx, map[string]interface{}{"_id": shippingMark.ShippingProvider}, nil)
		if err != nil {
			s.Logger.Error("Failed to find shipping provider", "correlationID", correlationID, "error", err)
			return err
		}
	}

	_, err = s.ShippingMarkRepository.Update(ctx, map[string]interface{}{"_id": shippingMarkID}, shippingMark)
	if err != nil {
		s.Logger.Error("Failed to update shipping mark", "correlationID", correlationID, "error", err)
		return err
	}

	s.Logger.Info("Shipping mark updated successfully", "correlationID", correlationID)

	return nil
}
