package services

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	validator "github.com/go-playground/validator/v10"
	"github.com/himdhiman/dashboard-backend/libs/cache"
	conflux_client "github.com/himdhiman/dashboard-backend/libs/conflux/pkg/client"
	"github.com/himdhiman/dashboard-backend/libs/logger"
	mongo_errors "github.com/himdhiman/dashboard-backend/libs/mongo/errors"
	mongo_models "github.com/himdhiman/dashboard-backend/libs/mongo/models"
	"github.com/himdhiman/dashboard-backend/libs/mongo/repository"
	"github.com/himdhiman/dashboard-backend/services/sentinel-service/models"
)

type UnicommerceService struct {
	Logger                  logger.ILogger
	Cache                   cache.Cacher
	ApiClient               *conflux_client.ConfluxAPIClient
	PurchaseOrderRepository *repository.Repository[models.PurchaseOrder]
}

func NewUnicommerceService(apiClient *conflux_client.ConfluxAPIClient, logger logger.ILogger, cache cache.Cacher,
	po_collections *mongo_models.MongoCollection) *UnicommerceService {

	purchaseOrderRepo := repository.Repository[models.PurchaseOrder]{Collection: po_collections}

	return &UnicommerceService{
		ApiClient:               apiClient,
		Cache:                   cache,
		Logger:                  logger,
		PurchaseOrderRepository: &purchaseOrderRepo,
	}
}

// CreatePurchaseOrder creates a new purchase order with an incremental order number
func (s *UnicommerceService) CreatePurchaseOrder(ctx context.Context, purchaseOrder *models.PurchaseOrder) error {
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

func (s *UnicommerceService) UpdatePurchaseOrder(ctx context.Context, poNumber string, updates map[string]interface{}) error {
	// Fetch the purchase order
	purchaseOrder, err := s.PurchaseOrderRepository.FindOne(ctx, map[string]interface{}{"poNumber": poNumber}, nil)
	if err != nil {
		s.Logger.Error("Error fetching purchase order", "error", err)
		return err
	}

	// Update the fields
	for fieldPath, value := range updates {
		if !isAllowedField(fieldPath) {
			return fmt.Errorf("field %s is not allowed to be updated", fieldPath)
		}

		err := setField(purchaseOrder, fieldPath, value)
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

func isAllowedField(fieldPath string) bool {

	var allowedFields = map[string]bool{
		"OrderStatus":           true,
		"TotalAmount":           true,
		"TentativeDispatchDate": true,
		"Remarks":               true,
	}

	var allowedProductFields = map[string]bool{
		"SkuCode":         true,
		"ImageURL":        true,
		"Quantity":        true,
		"CurrentRMBPrice": true,
		"Status":          true,
		"Remarks":         true,
		"ShippingMark":    true,
	}

	fields := strings.Split(fieldPath, ".")
	if len(fields) == 1 {
		return allowedFields[fields[0]]
	} else if len(fields) == 3 && fields[0] == "Products" {
		return allowedProductFields[fields[2]]
	}
	return false
}

// setField navigates through obj based on the dot-separated fieldPath.
// It supports nested fields and has special handling for "Products" where the second token is the SKU.
// If the SKU is new, a new product entry is created, its SkuCode is set, and it is appended to the purchase order.
func setField(obj interface{}, fieldPath string, value interface{}) error {
	fields := strings.Split(fieldPath, ".")
	// Start with the base object; we assume obj is a pointer.
	v := reflect.ValueOf(obj).Elem()

	// Process the field tokens one by one.
	for len(fields) > 0 {
		field := fields[0]
		fields = fields[1:]

		if field == "Products" {
			// Next token must be the SKU.
			if len(fields) < 2 {
				return fmt.Errorf("invalid field path for Products: %s", fieldPath)
			}
			sku := fields[0]
			// Consume the SKU token.
			fields = fields[1:]
			productsField := v.FieldByName("Products")
			if !productsField.IsValid() {
				return fmt.Errorf("no such field: Products")
			}
			if productsField.Kind() != reflect.Slice {
				return fmt.Errorf("products field is not a slice")
			}
			var product reflect.Value
			found := false
			// Search for an existing product with the given SKU.
			for i := 0; i < productsField.Len(); i++ {
				candidate := productsField.Index(i)
				if candidate.FieldByName("SkuCode").String() == sku {
					product = candidate
					found = true
					break
				}
			}
			if !found {
				// Create a new product instance.
				elemType := productsField.Type().Elem()
				newProduct := reflect.New(elemType).Elem()
				// Set the SkuCode on the new product.
				skuField := newProduct.FieldByName("SkuCode")
				if skuField.IsValid() && skuField.CanSet() && skuField.Kind() == reflect.String {
					skuField.SetString(sku)
				} else {
					return fmt.Errorf("cannot set SkuCode on new product for SKU %s", sku)
				}
				// Append the new product to the Products slice.
				newSlice := reflect.Append(productsField, newProduct)
				productsField.Set(newSlice)
				// Retrieve the newly added product.
				product = newSlice.Index(newSlice.Len() - 1)
			}
			// Now continue updating within the found or newly created product.
			v = product
		} else {
			// If no further tokens, then this field should be set.
			if len(fields) == 0 {
				f := v.FieldByName(field)
				if !f.IsValid() {
					return fmt.Errorf("no such field: %s in object", field)
				}
				if !f.CanSet() {
					return fmt.Errorf("cannot set field %s", field)
				}
				val := reflect.ValueOf(value)
				// Special handling for time.Time fields.
				if f.Type() == reflect.TypeOf(time.Time{}) {
					str, ok := value.(string)
					if !ok {
						return fmt.Errorf("expected string value for time field %s", field)
					}
					parsedTime, err := time.Parse(time.RFC3339, str)
					if err != nil {
						return fmt.Errorf("error parsing time for field %s: %v", field, err)
					}
					val = reflect.ValueOf(parsedTime)
				} else if f.Type() != val.Type() {
					return fmt.Errorf("provided value type didn't match field %s type: expected %s but got %s", field, f.Type(), val.Type())
				}
				f.Set(val)
				return nil
			} else {
				// Not the final field: move deeper into the object.
				v = v.FieldByName(field)
				if !v.IsValid() {
					return fmt.Errorf("no such field: %s in object", field)
				}
				// If v is a pointer, ensure it is non-nil.
				if v.Kind() == reflect.Ptr {
					if v.IsNil() {
						v.Set(reflect.New(v.Type().Elem()))
					}
					v = v.Elem()
				}
			}
		}
	}

	return nil
}

func (s *UnicommerceService) GetPurchaseOrders(ctx context.Context, poNumber string, pageNumber int, fieldsPerPage int) ([]*models.PurchaseOrder, int64, error) {
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
