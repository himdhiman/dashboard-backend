package services

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/go-playground/validator"
	constants "github.com/himdhiman/dashboard-backend/libs/constants"
	"github.com/himdhiman/dashboard-backend/libs/logger"
	"github.com/himdhiman/dashboard-backend/libs/mappers"
	"github.com/himdhiman/dashboard-backend/libs/mongo"
	mongo_errors "github.com/himdhiman/dashboard-backend/libs/mongo/errors"
	mongo_models "github.com/himdhiman/dashboard-backend/libs/mongo/models"
	"github.com/himdhiman/dashboard-backend/libs/mongo/repository"
	product_services "github.com/himdhiman/dashboard-backend/services/products-service/services"
	"github.com/himdhiman/dashboard-backend/services/purchaseOrder-service/dto"
	"github.com/himdhiman/dashboard-backend/services/purchaseOrder-service/models"
	"github.com/himdhiman/dashboard-backend/services/purchaseOrder-service/utils"
)

type PurchaseOrderService struct {
	Logger                          logger.ILogger
	Mapper                          *mappers.Mapper
	PurchaseOrderRepository         *repository.Repository[models.PurchaseOrder]
	PurchaseOrderProductsRepository *repository.Repository[models.PurchaseOrderProducts]
	ProductsService                 product_services.ProductsService
}

func NewPurchaseOrderService(logger logger.ILogger,
	mapper *mappers.Mapper,
	purchaseOrderRepository *repository.Repository[models.PurchaseOrder],
	purchaseOrderProductsRepository *repository.Repository[models.PurchaseOrderProducts],
	productsService product_services.ProductsService) *PurchaseOrderService {

	return &PurchaseOrderService{
		Logger:                          logger,
		Mapper:                          mapper,
		PurchaseOrderRepository:         purchaseOrderRepository,
		PurchaseOrderProductsRepository: purchaseOrderProductsRepository,
		ProductsService:                 productsService,
	}
}

// CreatePurchaseOrder creates a new purchase order with an incremental order number
func (s *PurchaseOrderService) CreatePurchaseOrder(ctx context.Context, purchaseOrderDTO *dto.CreatePurchaseOrderDTO) (*dto.CreatePurchaseOrderResponse, error) {
	// Fetch the last purchase order to determine the next order number
	correlationID := ctx.Value(constants.CorrelationID).(string)
	s.Logger.Info("Creating purchase order", "correlationID", correlationID)

	// Map the DTO to the model
	var purchaseOrder models.PurchaseOrder
	err := s.Mapper.Decode(purchaseOrderDTO, &purchaseOrder)
	if err != nil {
		s.Logger.Error("Error mapping DTO to model", "correlationID", correlationID, "error", err)
		return nil, err
	}

	// check if the vendor is valid or not
	isVendorValid, err := s.ProductsService.IsValidVendor(ctx, purchaseOrder.Vendor)
	if err != nil {
		s.Logger.Error("Error checking if vendor is valid", "correlationID", correlationID, "error", err)
		return nil, err
	}

	if !isVendorValid {
		s.Logger.Error("Invalid vendor", "correlationID", correlationID, "vendor", purchaseOrder.Vendor)
		return nil, fmt.Errorf("invalid vendor")
	}

	lastOrder, err := s.PurchaseOrderRepository.FindOne(ctx, nil, &mongo_models.FindOptions{
		Sort: map[string]interface{}{"poNumber": -1},
	})
	if err != nil && err != mongo_errors.ErrDocumentNotFound {
		s.Logger.Error("Error fetching last purchase order", "correlationID", correlationID, "error", err)
		return nil, err
	}

	// Determine the next order number
	var nextOrderNumber int
	if lastOrder != nil {
		re := regexp.MustCompile(`\d+$`)
		lastOrderNumberStr := re.FindString(lastOrder.PONumber)
		lastOrderNumber, err := strconv.Atoi(lastOrderNumberStr)
		if err != nil {
			s.Logger.Error("Error converting last order number to integer", "correlationID", correlationID, "error", err)
			return nil, err
		}
		nextOrderNumber = lastOrderNumber + 1
	} else {
		nextOrderNumber = 1
	}

	// Format the PO number as PO/(vendor)/(date)01
	vendor := purchaseOrder.Vendor
	date := time.Now().Format("20060102")
	purchaseOrder.PONumber = fmt.Sprintf("PO/%s/%s/%02d", vendor, date, nextOrderNumber)

	// Set the order date and initial status
	purchaseOrder.OrderDate = time.Now()
	purchaseOrder.UpdatedAt = time.Now()
	purchaseOrder.ShippingStatus = "pending"
	purchaseOrder.OrderStatus = "pending"

	purchaseOrder.Products = []mongo.ObjectID{}

	// validate the purchase order use validator v10
	validator := validator.New()
	err = validator.Struct(purchaseOrder)
	if err != nil {
		s.Logger.Error("Error validating purchase order", "correlationID", correlationID, "error", err)
		return nil, err
	}

	purchaseOrder.ID = mongo.NewObjectID()

	// Save the purchase order to the database
	_, err = s.PurchaseOrderRepository.Create(ctx, &purchaseOrder)
	if err != nil {
		s.Logger.Error("Error creating purchase order in DB", "correlationID", correlationID, "error", err)
		return nil, err
	}

	return &dto.CreatePurchaseOrderResponse{
		ID:       purchaseOrder.ID.Hex(),
		PONumber: purchaseOrder.PONumber,
	}, nil
}

// UpdatePurchaseOrder updates an existing purchase order
func (s *PurchaseOrderService) UpdatePurchaseOrder(ctx context.Context, poID string, updates map[string]interface{}) error {
	// Fetch the purchase order
	correlationID := ctx.Value(constants.CorrelationID).(string)
	s.Logger.Info("Updating purchase order", "correlationID", correlationID)

	purchaseOrder, err := s.PurchaseOrderRepository.FindOne(ctx, map[string]interface{}{"_id": poID}, nil)
	if err != nil {
		s.Logger.Error("Error fetching purchase order", "correlationID", correlationID, "error", err)
		return err
	}

	// Update the fields
	for fieldPath, value := range updates {
		err := utils.SetField(purchaseOrder, fieldPath, value, utils.AllowedFields)
		if err != nil {
			s.Logger.Error("Error setting field", "correlationID", correlationID, "fieldPath", fieldPath, "error", err)
			return err
		}
	}

	// Update the updatedAt field
	purchaseOrder.UpdatedAt = time.Now()

	validator := validator.New()
	err = validator.Struct(purchaseOrder)
	if err != nil {
		s.Logger.Error("Error validating purchase order", "correlationID", correlationID, "error", err)
		return err
	}

	// Save the updated purchase order
	_, err = s.PurchaseOrderRepository.Update(ctx, map[string]interface{}{"_id": poID}, purchaseOrder)
	if err != nil {
		s.Logger.Error("Error updating purchase order in DB", "correlationID", correlationID, "error", err)
		return err
	}

	s.Logger.Info("Successfully updated purchase order", "correlationID", correlationID)

	return nil
}

func (s *PurchaseOrderService) ListPurchaseOrders(ctx context.Context, poNumber string, pageNumber int, fieldsPerPage int) ([]dto.ListPurchaseOrdersDTO, int64, error) {
	correlationID := ctx.Value(constants.CorrelationID).(string)
	s.Logger.Info("Getting purchase orders", "correlationID", correlationID)

	filter := map[string]interface{}{}
	if poNumber != "" {
		filter["poNumber"] = poNumber
	}

	// Ensure pageNumber is at least 1
	if pageNumber < 1 {
		pageNumber = 1
	}

	purchaseOrders, err := s.PurchaseOrderRepository.Find(ctx, filter, &mongo_models.FindOptions{
		Limit: int64(fieldsPerPage),
		Skip:  int64((pageNumber - 1) * fieldsPerPage),
	})

	if err != nil {
		s.Logger.Error("Error fetching purchase orders", "correlationID", correlationID, "error", err)
		return nil, 0, err
	}

	cnt, err := s.PurchaseOrderRepository.Count(ctx, filter)
	if err != nil {
		s.Logger.Error("Error fetching purchase orders count", "correlationID", correlationID, "error", err)
		return nil, 0, err
	}

	var purchaseOrdersDTO []dto.ListPurchaseOrdersDTO

	err = s.Mapper.Decode(purchaseOrders, &purchaseOrdersDTO)
	if err != nil {
		s.Logger.Error("Error decoding purchase orders", "correlationID", correlationID, "error", err)
		return nil, 0, err
	}

	return purchaseOrdersDTO, cnt, nil
}

func (s *PurchaseOrderService) DeletePurchaseOrder(ctx context.Context, poID string) error {
	correlationID := ctx.Value(constants.CorrelationID).(string)
	s.Logger.Info("Deleting purchase order", "correlationID", correlationID)

	_, err := s.PurchaseOrderRepository.Delete(ctx, map[string]interface{}{"_id": poID})
	if err != nil {
		s.Logger.Error("Error deleting purchase order", "correlationID", correlationID, "error", err)
		return err
	}

	return nil
}

func (s *PurchaseOrderService) GetPurchaseOrder(ctx context.Context, poID string) (*dto.PurchaseOrderDTO, error) {
	correlationID := ctx.Value(constants.CorrelationID).(string)
	s.Logger.Info("Getting purchase order", "correlationID", correlationID)

	purchaseOrder, err := s.PurchaseOrderRepository.FindOne(ctx, map[string]interface{}{"_id": poID}, nil)
	if err != nil {
		s.Logger.Error("Error fetching purchase order", "correlationID", correlationID, "error", err)
		return nil, err
	}

	var purchaseOrderDTO dto.PurchaseOrderDTO
	err = s.Mapper.Decode(purchaseOrder, &purchaseOrderDTO)
	if err != nil {
		s.Logger.Error("Error mapping model to DTO", "correlationID", correlationID, "error", err)
		return nil, err
	}

	return &purchaseOrderDTO, nil
}

// AddProductToPurchaseOrder adds a product to an existing purchase order
func (s *PurchaseOrderService) AddProductToPurchaseOrder(ctx context.Context, purchaseOrderID string, req *dto.CreatePurchaseOrderProductDTO) (*dto.CreatePurchaseOrderProductResponse, error) {
	// Fetch the purchase order
	correlationID := ctx.Value(constants.CorrelationID).(string)
	s.Logger.Info("Adding product to purchase order", "correlationID", correlationID)

	// Map the DTO to the model
	var product models.PurchaseOrderProducts
	err := s.Mapper.Decode(req, &product)
	if err != nil {
		s.Logger.Error("Error mapping DTO to model", "correlationID", correlationID, "error", err)
		return nil, err
	}

	// Fetch the purchase order
	purchaseOrder, err := s.PurchaseOrderRepository.FindOne(ctx, map[string]interface{}{"_id": purchaseOrderID}, nil)
	if err != nil {
		s.Logger.Error("Error fetching purchase order", "correlationID", correlationID, "error", err)
		return nil, err
	}

	// Check if the product is already in the purchase order
	for _, existingProduct := range purchaseOrder.Products {
		if existingProduct.Hex() == product.ProductID {
			s.Logger.Error("Product already exists in purchase order", "correlationID", correlationID, "productID", product.ProductID)
			return nil, fmt.Errorf("product with ID %s already exists in purchase order %s", product.ProductID, purchaseOrderID)
		}
	}

	// Add the product to the purchase order
	product.ID = mongo.NewObjectID()
	purchaseOrder.Products = append(purchaseOrder.Products, product.ID)

	// Update the updatedAt field
	purchaseOrder.UpdatedAt = time.Now()

	// Save the updated purchase order
	_, err = s.PurchaseOrderRepository.Update(ctx, map[string]interface{}{"_id": purchaseOrderID}, purchaseOrder)
	if err != nil {
		s.Logger.Error("Error updating purchase order in DB", "correlationID", correlationID, "error", err)
		return nil, err
	}

	product.CreatedAt = time.Now()
	product.UpdatedAt = time.Now()

	// Save the product to the database
	_, err = s.PurchaseOrderProductsRepository.Create(ctx, &product)
	if err != nil {
		s.Logger.Error("Error creating product in DB", "correlationID", correlationID, "error", err)
		return nil, err
	}

	return &dto.CreatePurchaseOrderProductResponse{
		ID: product.ID.Hex(),
	}, nil
}

func (s *PurchaseOrderService) UpdatePurchaseOrderProduct(ctx context.Context, productID string, updates map[string]interface{}) error {
	// Fetch the product
	correlationID := ctx.Value(constants.CorrelationID).(string)
	s.Logger.Info("Updating purchase order product", "correlationID", correlationID)

	product, err := s.PurchaseOrderProductsRepository.FindOne(ctx, map[string]interface{}{"_id": productID}, nil)
	if err != nil {
		s.Logger.Error("Error fetching product", "correlationID", correlationID, "error", err)
		return err
	}

	// Update the fields
	for fieldPath, value := range updates {
		err := utils.SetField(product, fieldPath, value, utils.AllowedProductFields)
		if err != nil {
			s.Logger.Error("Error setting field", "correlationID", correlationID, "fieldPath", fieldPath, "error", err)
			return err
		}
	}

	// Update the updatedAt field
	product.UpdatedAt = time.Now()

	// validate the purchase order use validator v10
	validator := validator.New()
	err = validator.Struct(product)
	if err != nil {
		s.Logger.Error("Error validating purchase order", "correlationID", correlationID, "error", err)
		return err
	}

	// Save the updated product
	_, err = s.PurchaseOrderProductsRepository.Update(ctx, map[string]interface{}{"_id": productID}, product)
	if err != nil {
		s.Logger.Error("Error updating product in DB", "correlationID", correlationID, "error", err)
		return err
	}

	return nil
}

// func (s *PurchaseOrderService) updatePurchaseOrderStatus(ctx context.Context, purchaseOrder *models.PurchaseOrder) error {
// 	correlationID := ctx.Value(constants.CorrelationID).(string)
// 	s.Logger.Info("Updating purchase order status", "correlationID", correlationID)

// 	// Initialize counters for SKU statuses
// 	finalCount := 0
// 	pendingCount := 0

// 	// Count the number of SKUs with each status
// 	for _, sku := range purchaseOrder.Products {
// 		if sku.Status == "finalized" {
// 			finalCount++
// 		} else if sku.Status == "pending" {
// 			pendingCount++
// 		}
// 	}

// 	totalSkus := len(purchaseOrder.Products)

// 	// Determine overall PO status based on SKU statuses
// 	var newStatus string
// 	if finalCount == totalSkus {
// 		newStatus = "finalized"
// 	} else if pendingCount == totalSkus {
// 		newStatus = "pending"
// 	} else {
// 		newStatus = "partially_pending"
// 	}

// 	// Update the PO status if it has changed
// 	if purchaseOrder.OrderStatus != newStatus {
// 		purchaseOrder.OrderStatus = newStatus
// 	}

// 	return nil
// }

// func (s *PurchaseOrderService) updateShippingStatus(ctx context.Context, purchaseOrder *models.PurchaseOrder) error {
// 	correlationID := ctx.Value(constants.CorrelationID).(string)
// 	s.Logger.Info("Updating purchase order shipping status", "correlationID", correlationID)

// 	// Initialize counters for shipping mark status
// 	totalSkus := len(purchaseOrder.Products)
// 	skusWithShippingMark := 0

// 	// Count SKUs with shipping marks
// 	for _, product := range purchaseOrder.Products {
// 		if product.ShippingMark != "" {
// 			skusWithShippingMark++
// 		}
// 	}

// 	// Determine shipping status based on shipping marks
// 	var newStatus string
// 	if skusWithShippingMark == 0 {
// 		newStatus = "pending"
// 	} else if skusWithShippingMark == totalSkus {
// 		newStatus = "complete"
// 	} else {
// 		newStatus = "partly_shipped"
// 	}

// 	// Update the shipping status if it has changed
// 	if purchaseOrder.ShippingStatus != newStatus {
// 		purchaseOrder.ShippingStatus = newStatus
// 		purchaseOrder.UpdatedAt = time.Now()
// 	}

// 	return nil
// }

// func (s *PurchaseOrderService) DeleteProductFromPurchaseOrder(ctx context.Context, poNumber string, skuCode string) error {
// 	correlationID := ctx.Value(constants.CorrelationID).(string)
// 	s.Logger.Info("Deleting product from purchase order", "correlationID", correlationID)

// 	// Fetch the purchase order
// 	purchaseOrder, err := s.PurchaseOrderRepository.FindOne(ctx, map[string]interface{}{"poNumber": poNumber}, nil)
// 	if err != nil {
// 		s.Logger.Error("Error fetching purchase order", "correlationID", correlationID, "error", err)
// 		return err
// 	}

// 	// Find the product in the purchase order
// 	productIndex := -1
// 	for i, product := range purchaseOrder.Products {
// 		if product.SkuCode == skuCode {
// 			productIndex = i
// 			break
// 		}
// 	}

// 	if productIndex == -1 {
// 		s.Logger.Error("Product not found in purchase order", "correlationID", correlationID, "skuCode", skuCode)
// 		return fmt.Errorf("product with SKU %s not found in purchase order %s", skuCode, poNumber)
// 	}

// 	// Remove the product from the purchase order
// 	purchaseOrder.Products = append(purchaseOrder.Products[:productIndex], purchaseOrder.Products[productIndex+1:]...)

// 	// Update the updatedAt field
// 	purchaseOrder.UpdatedAt = time.Now()

// 	// Save the updated purchase order
// 	_, err = s.PurchaseOrderRepository.Update(ctx, map[string]interface{}{"poNumber": poNumber}, purchaseOrder)
// 	if err != nil {
// 		s.Logger.Error("Error updating purchase order in DB", "correlationID", correlationID, "error", err)
// 		return err
// 	}

// 	return nil
// }
