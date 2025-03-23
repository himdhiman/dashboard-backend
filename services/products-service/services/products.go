package services

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/himdhiman/dashboard-backend/libs/cache"
	conflux_client "github.com/himdhiman/dashboard-backend/libs/conflux/pkg/client"
	conflux_models "github.com/himdhiman/dashboard-backend/libs/conflux/pkg/models"
	"github.com/himdhiman/dashboard-backend/libs/constants"
	"github.com/himdhiman/dashboard-backend/libs/logger"
	mongo_models "github.com/himdhiman/dashboard-backend/libs/mongo/models"
	"github.com/himdhiman/dashboard-backend/libs/mongo/repository"
	products_constants "github.com/himdhiman/dashboard-backend/services/products-service/constants"
	"github.com/himdhiman/dashboard-backend/services/products-service/models"
)

type ProductsService struct {
	ServiceCode        string
	Logger             logger.ILogger
	Cache              cache.Cacher
	ApiClient          *conflux_client.ConfluxAPIClient
	GoogleSheetService *GoogleSheetsService
	ProductsRepository *repository.Repository[models.Product]
}

func NewProductsService(apiClient *conflux_client.ConfluxAPIClient, sheetService *GoogleSheetsService, logger logger.ILogger, cache cache.Cacher,
	productsCollection *mongo_models.MongoCollection) *ProductsService {

	productsRepo := repository.Repository[models.Product]{Collection: productsCollection}

	return &ProductsService{
		ServiceCode:        products_constants.UNICOM_API_CODE,
		ApiClient:          apiClient,
		Cache:              cache,
		GoogleSheetService: sheetService,
		Logger:             logger,
		ProductsRepository: &productsRepo,
	}
}

func (s *ProductsService) IsValidVendor(ctx context.Context, vendorID string) (bool, error) {
	if vendorID == "" {
		return false, errors.New("Vendor ID cannot be empty")
	}

	filter := map[string]interface{}{
		"primaryVendor": vendorID,
	}
	productsCount, err := s.ProductsRepository.Count(ctx, filter)
	if err != nil {
		s.Logger.Error("Error fetching products", "error", err)
		return false, err
	}

	if productsCount == 0 {
		return false, nil
	}

	return true, nil
}

func (s *ProductsService) GetProductIDBySKUVendor(ctx context.Context, skuCode string, vendorID string) (string, error) {
	if skuCode == "" {
		return "", errors.New("SKU code cannot be empty")
	}

	if vendorID == "" {
		return "", errors.New("Vendor ID cannot be empty")
	}

	filter := map[string]interface{}{
		"skuCode":       skuCode,
		"primaryVendor": vendorID,
	}
	products, err := s.ProductsRepository.Find(ctx, filter)
	if err != nil {
		s.Logger.Error("Error fetching products", "error", err)
		return "", err
	}

	if len(products) == 0 {
		return "", nil
	}

	return products[0].ID.Hex(), nil
}

func (s *ProductsService) UpdateInventoryFromGoogleSheet(ctx context.Context) error {
	correlationID, ok := ctx.Value(constants.CorrelationID).(string)
	if !ok {
		s.Logger.Error("Correlation ID not found in context")
		return errors.New("correlation ID not found in context")
	}

	s.Logger.Info("Starting inventory update from Google Sheet", "correlationID", correlationID)

	// Read SKUs from Google Sheet
	s.Logger.Info("Fetching data from Google Sheet", "correlationID", correlationID)
	sheetData, err := s.GoogleSheetService.FetchGoogleSheetData(ctx)
	if err != nil {
		s.Logger.Error("Error reading SKUs from Google Sheet", "error", err, "correlationID", correlationID)
		return err
	}
	s.Logger.Info("Successfully fetched data from Google Sheet", "rowCount", len(sheetData), "correlationID", correlationID)

	// Extract SKUs from the Google Sheet data
	var skus []string
	for _, row := range sheetData {
		if len(row) > 1 {
			skus = append(skus, row["SKU"].(string))
		}
	}
	s.Logger.Info("Extracted SKUs from Google Sheet data", "skuCount", len(skus), "correlationID", correlationID)

	// Fetch inventory snapshot for all SKUs
	s.Logger.Info("Fetching inventory snapshot for SKUs", "correlationID", correlationID)
	inventorySnapshot, err := s.GetInventorySnapshot(ctx, skus)
	if err != nil {
		s.Logger.Error("Error fetching inventory snapshot", "error", err, "correlationID", correlationID)
		return err
	}
	s.Logger.Info("Successfully fetched inventory snapshot", "snapshotCount", len(inventorySnapshot), "correlationID", correlationID)

	// Update inventory in sheetData and save it back to Google Sheet
	s.Logger.Info("Updating Google Sheet data with inventory snapshot", "correlationID", correlationID)
	for i, row := range sheetData {
		if len(row) > 1 {
			sku := row["SKU"].(string)
			if inventory, ok := inventorySnapshot[sku]; ok {
				sheetData[i]["Quantity"] = inventory
				sheetData[i]["Last Updated"] = time.Now().Format("2006-01-02 15:04:05")
			} else {
				s.Logger.Warn("No inventory data found for SKU", "SKU", sku, "correlationID", correlationID)
			}
		}
	}

	s.Logger.Info("Saving updated data back to Google Sheet", "correlationID", correlationID)
	err = s.GoogleSheetService.UpdateGoogleSheet(ctx, sheetData)
	if err != nil {
		s.Logger.Error("Error updating Google Sheet", "error", err, "correlationID", correlationID)
		return err
	}
	s.Logger.Info("Successfully updated Google Sheet", "correlationID", correlationID)

	return nil
}

// Create a function which will make a post request to unicommerce and get the inventrory snapshot, we will provide the list of SKUs
func (s *ProductsService) GetInventorySnapshot(ctx context.Context, skus []string) (map[string]int, error) {
	correlationID, ok := ctx.Value(constants.CorrelationID).(string)
	if !ok {
		s.Logger.Error("Correlation ID not found in context")
		return nil, errors.New("correlation ID not found in context")
	}

	payload := map[string]interface{}{
		"itemTypeSKUs": skus,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		s.Logger.Error("Error encoding payload for inventory snapshot request", "error", err, "correlationID", correlationID)
		return nil, err
	}

	headers := map[string]string{
		"Content-Type": "application/json",
		"Facility":     "Salty",
	}

	resp, err := s.ApiClient.DoRequest(ctx, &conflux_models.APIRequest{
		ApiCode: products_constants.API_CODE_GET_INVENTORY_SNAPSHOT,
		Headers: headers,
		Body:    strings.NewReader(string(payloadBytes)),
	})

	if resp.StatusCode != http.StatusOK {
		s.Logger.Error("Error fetching inventory snapshot", "status", resp.StatusCode, "correlationID", correlationID)
		return nil, err
	}

	var responseData struct {
		InventorySnapshots []struct {
			ItemTypeSKU string `json:"itemTypeSKU"`
			Inventory   int    `json:"inventory"`
		} `json:"inventorySnapshots"`
	}

	err = json.Unmarshal(resp.Body, &responseData)
	if err != nil {
		s.Logger.Error("Error decoding response body", "error", err, "correlationID", correlationID)
		return nil, err
	}

	inventoryMap := make(map[string]int)
	for _, snapshot := range responseData.InventorySnapshots {
		inventoryMap[snapshot.ItemTypeSKU] = snapshot.Inventory
	}

	s.Logger.Info("Completed GetInventorySnapshot", "correlationID", correlationID)
	return inventoryMap, nil
}

func (s *ProductsService) GetProducts(ctx context.Context, skuCode string, pageNumber int, fieldsPerPage int) ([]*models.Product, int64, error) {
	filter := map[string]interface{}{}
	if skuCode != "" {
		filter["skuCode"] = skuCode
	}

	// Ensure pageNumber is at least 1
	if pageNumber < 1 {
		pageNumber = 1
	}

	products, err := s.ProductsRepository.Find(ctx, filter, &mongo_models.FindOptions{
		Limit: int64(fieldsPerPage),
		Skip:  int64((pageNumber - 1) * fieldsPerPage),
	})

	if err != nil {
		s.Logger.Error("Error fetching products", "error", err)
		return nil, 0, err
	}

	cnt, err := s.ProductsRepository.Count(ctx, filter)
	if err != nil {
		s.Logger.Error("Error fetching products Count", "error", err)
		return nil, 0, err
	}

	return products, cnt, nil
}

// Create a function to fetch the product by SKU code or by name with partial matching
func (s *ProductsService) SearchProduct(ctx context.Context, skuCode string, name string) ([]*models.Product, error) {
	filter := map[string]interface{}{}
	if skuCode != "" {
		filter["skuCode"] = map[string]interface{}{"$regex": skuCode, "$options": "i"}
	}
	if name != "" {
		filter["name"] = map[string]interface{}{"$regex": name, "$options": "i"}
	}
	products, err := s.ProductsRepository.Find(ctx, filter)
	if err != nil {
		s.Logger.Error("Error fetching products", "error", err)
		return nil, err
	}
	return products, nil
}
