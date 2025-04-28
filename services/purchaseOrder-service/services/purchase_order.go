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
	"github.com/himdhiman/dashboard-backend/libs/task"
	product_services "github.com/himdhiman/dashboard-backend/services/products-service/services"
	"github.com/himdhiman/dashboard-backend/services/purchaseOrder-service/dto"
	"github.com/himdhiman/dashboard-backend/services/purchaseOrder-service/models"
	"github.com/himdhiman/dashboard-backend/services/purchaseOrder-service/utils"
	shipping_services "github.com/himdhiman/dashboard-backend/services/shipping-service/services"
)

type PurchaseOrderService struct {
	Logger                          logger.ILogger
	Mapper                          *mappers.Mapper
	TaskManager                     *task.TaskManager
	PurchaseOrderRepository         *repository.Repository[models.PurchaseOrder]
	PurchaseOrderProductsRepository *repository.Repository[models.PurchaseOrderProducts]
	ProductsService                 product_services.ProductsService
	ShippingService                 shipping_services.ShippingService
}

func NewPurchaseOrderService(logger logger.ILogger,
	mapper *mappers.Mapper,
	taskManager *task.TaskManager,
	purchaseOrderRepository *repository.Repository[models.PurchaseOrder],
	purchaseOrderProductsRepository *repository.Repository[models.PurchaseOrderProducts],
	productsService product_services.ProductsService,
	shippingService shipping_services.ShippingService) *PurchaseOrderService {

	return &PurchaseOrderService{
		Logger:                          logger,
		Mapper:                          mapper,
		TaskManager:                     taskManager,
		PurchaseOrderRepository:         purchaseOrderRepository,
		PurchaseOrderProductsRepository: purchaseOrderProductsRepository,
		ProductsService:                 productsService,
		ShippingService:                 shippingService,
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
		err := constants.SetField(purchaseOrder, fieldPath, value, utils.AllowedFields)
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

func (s *PurchaseOrderService) ListPurchaseOrders(ctx context.Context, poNumber string, pageNumber, fieldsPerPage int) ([]dto.ListPurchaseOrdersDTO, int64, error) {
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

	err = s.Mapper.DecodeWithCustomHook(purchaseOrders, &purchaseOrdersDTO, mappers.DecodeObjectIDHookFunc(), mappers.EncodeTimeToStringHookFunc())
	if err != nil {
		s.Logger.Error("Error decoding purchase orders", "correlationID", correlationID, "error", err)
		return nil, 0, err
	}

	return purchaseOrdersDTO, cnt, nil
}

func (s *PurchaseOrderService) DeletePurchaseOrder(ctx context.Context, poID string) error {
	correlationID := ctx.Value(constants.CorrelationID).(string)
	s.Logger.Info("Deleting purchase order", "correlationID", correlationID)

	// Fetch the purchase order
	po, err := s.PurchaseOrderRepository.FindOne(ctx, map[string]interface{}{"_id": poID}, nil)
	if err != nil {
		s.Logger.Error("Error fetching purchase order", "correlationID", correlationID, "error", err)
		return err
	}

	// delete all the associated PO Products
	for _, productID := range po.Products {
		_, err = s.PurchaseOrderProductsRepository.Delete(ctx, map[string]interface{}{"_id": productID.Hex()})
		if err != nil {
			s.Logger.Error("Error deleting purchase order product", "correlationID", correlationID, "error", err)
			return err
		}
	}

	_, err = s.PurchaseOrderRepository.Delete(ctx, map[string]interface{}{"_id": poID})
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
	err = s.Mapper.DecodeWithCustomHook(purchaseOrder, &purchaseOrderDTO, mappers.DecodeObjectIDHookFunc(), mappers.EncodeTimeToStringHookFunc())
	if err != nil {
		s.Logger.Error("Error mapping model to DTO", "correlationID", correlationID, "error", err)
		return nil, err
	}

	// Manually convert time.Time fields to string
	purchaseOrderDTO.OrderDate = purchaseOrder.OrderDate.Format(time.RFC3339)
	purchaseOrderDTO.TentativeDispatchDate = purchaseOrder.TentativeDispatchDate.Format(time.RFC3339)

	// Fetch the products associated with the purchase order, need to do for loop over the products and fetch them
	var purchaseOrderProducts []dto.PurchaseOrderProductDTO
	for _, productID := range purchaseOrder.Products {
		purchaseOrderProduct, err := s.PurchaseOrderProductsRepository.FindOne(ctx, map[string]interface{}{"_id": productID.Hex()}, nil)
		if err != nil {
			s.Logger.Error("Error fetching purchase order product", "correlationID", correlationID, "error", err)
			return nil, err
		}
		var purchaseOrderProductDTO dto.PurchaseOrderProductDTO
		err = s.Mapper.DecodeWithCustomHook(purchaseOrderProduct, &purchaseOrderProductDTO, mappers.DecodeObjectIDHookFunc(), mappers.EncodeTimeToStringHookFunc())
		if err != nil {
			s.Logger.Error("Error mapping model to DTO", "correlationID", correlationID, "error", err)
			return nil, err
		}

		product, err := s.ProductsService.GetProductByID(ctx, purchaseOrderProduct.ProductID)
		if err != nil {
			s.Logger.Error("Error fetching product", "correlationID", correlationID, "error", err)
			return nil, err
		}

		purchaseOrderProductDTO.SKUCode = product.SKUCode
		purchaseOrderProductDTO.ImageURL = product.ImageURL

		// Manually convert time.Time fields to string
		purchaseOrderProductDTO.OrderDate = purchaseOrderProduct.OrderDate.Format(time.RFC3339)

		purchaseOrderProducts = append(purchaseOrderProducts, purchaseOrderProductDTO)
	}

	purchaseOrderDTO.Products = purchaseOrderProducts

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

func (s *PurchaseOrderService) UpdatePurchaseOrderProduct(ctx context.Context, poID, productID string, updates map[string]interface{}) error {
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
		err := constants.SetField(product, fieldPath, value, utils.AllowedProductFields)
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

	// Check if the updates have the "status" field, spawn a task to update the purchase order status
	if _, ok := updates["Status"]; ok {
		taskParams := map[string]interface{}{
			"purchaseOrderID": poID,
			"field":           "status",
		}

		_, err := s.spawnUpdateTask(ctx, "UpdatePurchaseOrderStatus", taskParams)
		if err != nil {
			s.Logger.Error("Error spawning task to update purchase order status", "correlationID", correlationID, "error", err)
			return err
		}
	}

	// Check if the updates have the "shippingMark" field, spawn a task to update the purchase order shipping status
	if _, ok := updates["ShippingMark"]; ok {
		taskParams := map[string]interface{}{
			"purchaseOrderID": poID,
			"field":           "shippingMark",
		}

		_, err := s.spawnUpdateTask(ctx, "UpdatePurchaseOrderShippingStatus", taskParams)
		if err != nil {
			s.Logger.Error("Error spawning task to update purchase order shipping status", "correlationID", correlationID, "error", err)
			return err
		}

		id, err := s.ShippingService.CreateShippingMark(ctx, product.ShippingMark)
		if err != nil {
			s.Logger.Error("Error creating shipping mark", "correlationID", correlationID, "error", err)
			return err
		}

		s.Logger.Info("Shipping mark created successfully", "correlationID", correlationID, "id", id)
	}

	return nil
}

func (s *PurchaseOrderService) spawnUpdateTask(ctx context.Context, taskType string, params map[string]interface{}) (string, error) {
	// Create a new context with a longer timeout
	taskCtx, _ := context.WithTimeout(context.Background(), 30*time.Second)

	correlationID := ctx.Value(constants.CorrelationID).(string)
	s.Logger.Info("Spawning background task", "correlationID", correlationID, "taskType", taskType)

	// Use taskCtx for MongoDB operations
	taskID, err := s.TaskManager.RunTask(taskType, params, func(params map[string]interface{}) (interface{}, error) {
		purchaseOrderID, ok := params["purchaseOrderID"].(string)
		if !ok {
			return nil, fmt.Errorf("invalid purchaseOrderID in task params")
		}

		field, ok := params["field"].(string)
		if !ok {
			return nil, fmt.Errorf("invalid field in task params")
		}

		// Fetch the purchase order using the new context
		purchaseOrder, err := s.PurchaseOrderRepository.FindOne(taskCtx, map[string]interface{}{"_id": purchaseOrderID}, nil)
		if err != nil {
			return nil, fmt.Errorf("error fetching purchase order: %v", err)
		}

		// Update the purchase order status or shipping status
		if field == "status" {
			err = s.updatePurchaseOrderStatus(taskCtx, purchaseOrder)
		} else if field == "shippingMark" {
			err = s.updateShippingStatus(taskCtx, purchaseOrder)
		}

		if err != nil {
			return nil, fmt.Errorf("error updating purchase order: %v", err)
		}

		return "Task completed successfully", nil
	})

	if err != nil {
		s.Logger.Error("Error spawning task", "correlationID", correlationID, "taskType", taskType, "error", err)
		return "", err
	}

	s.Logger.Info("Task spawned successfully", "correlationID", correlationID, "taskID", taskID)
	return taskID, nil
}

func (s *PurchaseOrderService) updatePurchaseOrderStatus(ctx context.Context, purchaseOrder *models.PurchaseOrder) error {
	correlationID, ok := ctx.Value(constants.CorrelationID).(string)
	if !ok {
		correlationID = "unknown"
	}
	s.Logger.Info("Updating purchase order status", "correlationID", correlationID)

	// Initialize counters for SKU statuses
	finalCount := 0
	pendingCount := 0

	options := &mongo_models.FindOptions{
		Projection: map[string]interface{}{"_id": 1, "status": 1}, // Fetch only required fields
	}

	// Count the number of SKUs with each status
	for _, productID := range purchaseOrder.Products {
		product, err := s.PurchaseOrderProductsRepository.FindOne(ctx, map[string]interface{}{"_id": productID.Hex()}, options)
		if err != nil {
			return err
		}
		if product.Status == "finalized" {
			finalCount++
		} else if product.Status == "pending" {
			pendingCount++
		}
	}

	totalSkus := len(purchaseOrder.Products)

	// Determine overall PO status based on SKU statuses
	var newStatus string
	if finalCount == totalSkus {
		newStatus = "finalized"
	} else if pendingCount == totalSkus {
		newStatus = "pending"
	} else {
		newStatus = "partially_pending"
	}

	// Update the PO status if it has changed
	if purchaseOrder.OrderStatus != newStatus {
		purchaseOrder.OrderStatus = newStatus
	}

	// validate the purchase order use validator v10
	validator := validator.New()
	err := validator.Struct(purchaseOrder)
	if err != nil {
		return err
	}

	// save the purchase order
	_, err = s.PurchaseOrderRepository.Update(ctx, map[string]interface{}{"_id": purchaseOrder.ID.Hex()}, purchaseOrder)
	if err != nil {
		return err
	}

	return nil
}

func (s *PurchaseOrderService) updateShippingStatus(ctx context.Context, purchaseOrder *models.PurchaseOrder) error {
	correlationID, ok := ctx.Value(constants.CorrelationID).(string)
	if !ok {
		correlationID = "unknown"
	}
	s.Logger.Info("Updating purchase order shipping status", "correlationID", correlationID)

	// Initialize counters for shipping mark status
	totalSkus := len(purchaseOrder.Products)
	skusWithShippingMark := 0

	options := &mongo_models.FindOptions{
		Projection: map[string]interface{}{"_id": 1, "shippingMark": 1}, // Fetch only required fields
	}

	// Count SKUs with shipping marks
	for _, productID := range purchaseOrder.Products {
		product, err := s.PurchaseOrderProductsRepository.FindOne(ctx, map[string]interface{}{"_id": productID.Hex()}, options)
		if err != nil {
			return err
		}
		if product.ShippingMark != "" {
			skusWithShippingMark++
		}
	}

	// Determine shipping status based on shipping marks
	var newStatus string
	if skusWithShippingMark == 0 {
		newStatus = "pending"
	} else if skusWithShippingMark == totalSkus {
		newStatus = "complete"
	} else {
		newStatus = "partly_shipped"
	}

	// Update the shipping status if it has changed
	if purchaseOrder.ShippingStatus != newStatus {
		purchaseOrder.ShippingStatus = newStatus
		purchaseOrder.UpdatedAt = time.Now()
	}

	// validate the purchase order use validator v10
	validate := validator.New()
	err := validate.Struct(purchaseOrder)
	if err != nil {
		return err
	}

	// save the purchase order
	_, err = s.PurchaseOrderRepository.Update(ctx, map[string]interface{}{"_id": purchaseOrder.ID.Hex()}, purchaseOrder)
	if err != nil {
		return err
	}

	return nil
}

func (s *PurchaseOrderService) DeleteProductFromPurchaseOrder(ctx context.Context, productID string) error {
	correlationID := ctx.Value(constants.CorrelationID).(string)
	s.Logger.Info("Deleting product from purchase order", "correlationID", correlationID)

	_, err := s.PurchaseOrderProductsRepository.Delete(ctx, map[string]interface{}{"_id": productID})
	if err != nil {
		s.Logger.Error("Error deleting product from purchase order", "correlationID", correlationID, "error", err)
		return err
	}

	return nil
}
