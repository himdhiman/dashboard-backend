package services

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/go-playground/validator"
	"github.com/himdhiman/dashboard-backend/libs/logger"
	mongo_errors "github.com/himdhiman/dashboard-backend/libs/mongo/errors"
	mongo_models "github.com/himdhiman/dashboard-backend/libs/mongo/models"
	"github.com/himdhiman/dashboard-backend/libs/mongo/repository"
	"github.com/himdhiman/dashboard-backend/services/purchaseOrder-service/models"
	"github.com/himdhiman/dashboard-backend/services/purchaseOrder-service/utils"
)

type PurchaseOrderService struct {
	Logger                  logger.ILogger
	PurchaseOrderRepository *repository.Repository[models.PurchaseOrder]
}

func NewPurchaseOrderService(logger logger.ILogger,
	po_collections *mongo_models.MongoCollection) *PurchaseOrderService {

	purchaseOrderRepo := repository.Repository[models.PurchaseOrder]{Collection: po_collections}

	return &PurchaseOrderService{
		Logger:                  logger,
		PurchaseOrderRepository: &purchaseOrderRepo,
	}
}

// CreatePurchaseOrder creates a new purchase order with an incremental order number
func (s *PurchaseOrderService) CreatePurchaseOrder(ctx context.Context, purchaseOrder *models.PurchaseOrder) error {
	// Fetch the last purchase order to determine the next order number
	lastOrder, err := s.PurchaseOrderRepository.FindOne(ctx, nil, &mongo_models.FindOptions{
		Sort: map[string]interface{}{"poNumber": -1},
	})
	if err != nil && err != mongo_errors.ErrDocumentNotFound {
		s.Logger.Error("Error fetching last purchase order", "error", err)
		return err
	}

	// Determine the next order number
	var nextOrderNumber int
	if lastOrder != nil {
		re := regexp.MustCompile(`\d+$`)
		lastOrderNumberStr := re.FindString(lastOrder.PONumber)
		lastOrderNumber, err := strconv.Atoi(lastOrderNumberStr)
		if err != nil {
			s.Logger.Error("Error converting last order number to integer", "error", err)
			return err
		}
		nextOrderNumber = lastOrderNumber + 1
	} else {
		nextOrderNumber = 1
	}

	// Format the PO number as PO/(vendor)/(date)01
	vendor := purchaseOrder.Vendor
	date := time.Now().Format("20060102")
	purchaseOrder.PONumber = fmt.Sprintf("PO/%s/%s/%02d", vendor, date, nextOrderNumber)

	// Set the order date
	purchaseOrder.OrderDate = time.Now()
	purchaseOrder.UpdatedAt = time.Now()

	// Save the purchase order to the database
	_, err = s.PurchaseOrderRepository.Create(ctx, purchaseOrder)
	if err != nil {
		s.Logger.Error("Error creating purchase order in DB", "error", err)
		return err
	}

	return nil
}

func (s *PurchaseOrderService) UpdatePurchaseOrder(ctx context.Context, poNumber string, updates map[string]interface{}) error {
	// Fetch the purchase order
	purchaseOrder, err := s.PurchaseOrderRepository.FindOne(ctx, map[string]interface{}{"poNumber": poNumber}, nil)
	if err != nil {
		s.Logger.Error("Error fetching purchase order", "error", err)
		return err
	}

	// Update the fields
	for fieldPath, value := range updates {
		if !utils.IsAllowedField(fieldPath) {
			return fmt.Errorf("field %s is not allowed to be updated", fieldPath)
		}

		err := utils.SetField(purchaseOrder, fieldPath, value)
		if err != nil {
			s.Logger.Error("Error setting field", "fieldPath", fieldPath, "error", err)
			return err
		}
	}

	// Update the updatedAt field
	purchaseOrder.UpdatedAt = time.Now()

	// validate the purchase order use validator v10

	validator := validator.New()
	err = validator.Struct(purchaseOrder)
	if err != nil {
		s.Logger.Error("Error validating purchase order", "error", err)
		return err
	}

	// Save the updated purchase order
	_, err = s.PurchaseOrderRepository.Update(ctx, map[string]interface{}{"poNumber": poNumber}, purchaseOrder)
	if err != nil {
		s.Logger.Error("Error updating purchase order in DB", "error", err)
		return err
	}

	return nil
}

func (s *PurchaseOrderService) GetPurchaseOrders(ctx context.Context, poNumber string, pageNumber int, fieldsPerPage int) ([]*models.PurchaseOrder, int64, error) {
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
		s.Logger.Error("Error fetching purchase orders", "error", err)
		return nil, 0, err
	}

	cnt, err := s.PurchaseOrderRepository.Count(ctx, filter)
	if err != nil {
		s.Logger.Error("Error fetching purchase orders count", "error", err)
		return nil, 0, err
	}

	return purchaseOrders, cnt, nil
}
