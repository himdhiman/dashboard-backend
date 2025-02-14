package services

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator"
	"github.com/himdhiman/dashboard-backend/libs/cache"
	"github.com/himdhiman/dashboard-backend/libs/logger"
	mongo_errors "github.com/himdhiman/dashboard-backend/libs/mongo/errors"
	mongo_models "github.com/himdhiman/dashboard-backend/libs/mongo/models"
	"github.com/himdhiman/dashboard-backend/libs/mongo/repository"
	"github.com/himdhiman/dashboard-backend/services/sentinel-service/auth"
	"github.com/himdhiman/dashboard-backend/services/sentinel-service/constants"
	"github.com/himdhiman/dashboard-backend/services/sentinel-service/models"
)

type UnicommerceService struct {
	ServiceCode             string
	Logger                  logger.ILogger
	TokenManager            *auth.TokenManager
	GoogleSheetService      *GoogleSheetsService
	ProductsRepository      *repository.Repository[models.Product]
	PurchaseOrderRepository *repository.Repository[models.PurchaseOrder]
}

func NewUnicommerceService(tokenManager *auth.TokenManager, sheetService *GoogleSheetsService, logger logger.ILogger, productsCollection *mongo_models.MongoCollection, po_collections *mongo_models.MongoCollection) *UnicommerceService {

	productsRepo := repository.Repository[models.Product]{Collection: productsCollection}

	purchaseOrderRepo := repository.Repository[models.PurchaseOrder]{Collection: po_collections}

	return &UnicommerceService{
		ServiceCode:             constants.UNICOM_API_CODE,
		TokenManager:            tokenManager,
		GoogleSheetService:      sheetService,
		Logger:                  logger,
		ProductsRepository:      &productsRepo,
		PurchaseOrderRepository: &purchaseOrderRepo,
	}
}

func (s *UnicommerceService) fetchConfig(ctx context.Context, apiCode string) (string, string, string, int, error) {
	method, err := s.fetchFromCache(ctx, apiCode, constants.API_METHOD)
	if err != nil {
		s.Logger.Error("Error fetching method from cache", "error", err)
		return "", "", "", 0, err
	}

	baseURL, err := s.fetchFromCache(ctx, "", constants.BASE_URL)
	if err != nil {
		s.Logger.Error("Error fetching base URL from cache", "error", err)
		return "", "", "", 0, err
	}

	path, err := s.fetchFromCache(ctx, apiCode, constants.API_PATH)
	if err != nil {
		s.Logger.Error("Error fetching path from cache", "error", err)
		return "", "", "", 0, err
	}

	timeoutStr, err := s.fetchFromCache(ctx, apiCode, constants.API_TIMEOUT)
	if err != nil {
		s.Logger.Error("Error fetching timeout from cache", "error", err)
		return "", "", "", 0, err
	}

	timeout, cacheError := strconv.Atoi(timeoutStr)
	if cacheError != nil {
		s.Logger.Error("Error converting timeout to integer", "error", err)
		return "", "", "", 0, err
	}

	return method, baseURL, path, timeout, nil
}

type ExportJobResponse struct {
	Successful  bool     `json:"successful"`
	Message     string   `json:"message"`
	Errors      []string `json:"errors"`
	Warnings    string   `json:"warnings"`
	ExportJobID string   `json:"exportJobId"`
	JobCode     string   `json:"jobCode"`
}

type ExportJobPayload struct {
	ExportJobTypeName string      `json:"exportJobTypeName"`
	ExportColumns     []string    `json:"exportColums"`
	ExportFilters     interface{} `json:"exportFilters"`
	Frequency         string      `json:"frequency"`
	ReportName        string      `json:"reportName"`
}

type ExportJobStatusPayload struct {
	JobCode string `json:"jobCode"`
}

type ExportJobStatusResponse struct {
	Successful bool   `json:"successful"`
	Message    string `json:"message"`
	Status     string `json:"status"`
	FilePath   string `json:"filePath"`
}

func (s *UnicommerceService) CreateExportJob(ctx context.Context) (*ExportJobResponse, error) {
	method, baseURL, path, timeout, err := s.fetchConfig(ctx, constants.API_CODE_UNICOM_CREATE_JOB)
	if err != nil {
		return nil, err
	}

	fullURL := baseURL + path

	payload := &ExportJobPayload{
		ExportJobTypeName: "Item Master",
		ExportColumns:     []string{"skuCode", "itemName", "imageUrl", "type", "skuType", "itemType_Primary_Vendor"},
		ExportFilters:     nil,
		Frequency:         "ONETIME",
		ReportName:        time.Now().Format("2006-01-02 15:04:05"),
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		s.Logger.Error("Error encoding payload for token request", "error", err)
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, strings.NewReader(string(payloadBytes)))
	if err != nil {
		s.Logger.Error("Error creating request for FetchTokens", "error", err)
		return nil, err
	}

	s.TokenManager.AuthenticateRequest(ctx, req)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Facility", "salty")

	client := &http.Client{Timeout: time.Duration(timeout) * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		s.Logger.Error("Error making request to endpoint", "error", err)
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.Logger.Error("Error creating export job", "status", resp.StatusCode)
		return nil, err
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		s.Logger.Error("Error reading response body", "error", err)
		return nil, err
	}

	var exportJobResponse ExportJobResponse
	if err := json.Unmarshal(respBody, &exportJobResponse); err != nil {
		s.Logger.Error("Error decoding response body", "error", err)
		return nil, err
	}

	if !exportJobResponse.Successful {
		s.Logger.Error("Error creating export job", "message", exportJobResponse.Message)
		return nil, err
	}

	err = s.TokenManager.Cache.Set(ctx, s.ServiceCode+":"+constants.EXPORT_JOB_CODE, exportJobResponse.JobCode, 0)
	if err != nil {
		s.Logger.Error("Error setting export job code in cache", "error", err)
		return nil, err
	}

	return &exportJobResponse, nil
}

// check the job status and spin the task to read the data from csv and save it in mongo
func (s *UnicommerceService) CheckExportJobStatus(ctx context.Context) error {
	jobCode, err := s.fetchFromCache(ctx, constants.EXPORT_JOB_CODE, "")
	if err != nil {
		s.Logger.Error("Error fetching export job code from cache", "error", err)
		return err
	}

	s.Logger.Info("Checking export job status", "jobCode", jobCode)
	exportJobStatusResponse, cacheError := s.getExportJobStatus(ctx, jobCode)
	if cacheError != nil {
		s.Logger.Error("Error fetching export job status", "error", err)
		return err
	}
	if !exportJobStatusResponse.Successful {
		s.Logger.Error("Error fetching export job status", "message", exportJobStatusResponse.Message)
		return err
	}

	if exportJobStatusResponse.Status == "COMPLETE" {
		// we have the file url, we can now read the file iterate over each row and save it in mongo
		// we can use the file path to read the file

		// we have the aws file path, we can now read the file and save it in mongo
		fileURL := exportJobStatusResponse.FilePath
		resp, err := http.Get(fileURL)
		if err != nil {
			s.Logger.Error("Error downloading file from URL", "error", err)
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			s.Logger.Error("Error downloading file", "status", resp.StatusCode)
			return err
		}

		fileBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			s.Logger.Error("Error reading file content", "error", err)
			return err
		}

		// Assuming the file is a CSV, we can parse it
		r := csv.NewReader(bytes.NewReader(fileBytes))
		records, err := r.ReadAll()
		if err != nil {
			s.Logger.Error("Error parsing CSV file", "error", err)
			return err
		}

		for _, record := range records {
			if record[3] != "SIMPLE" {
				continue
			}

			var skuCode, name, imageURL, primaryVendor string

			skuCode = record[0]
			name = record[1]
			imageURL = record[2]
			primaryVendor = record[5]

			// check if already exists, the update the product
			// we can use the skuCode and primary vendor to check if the product already exists

			products, err := s.ProductsRepository.Find(ctx, map[string]interface{}{"skuCode": skuCode, "primaryVendor": primaryVendor})
			if err != nil {
				s.Logger.Error("Error fetching products", "error", err)
				return err
			}
			if len(products) > 0 {
				// if the product already exists, we update the product
				// we can use the skuCode and primary vendor to update the product
				_, err = s.ProductsRepository.Update(ctx, map[string]interface{}{"name": name, "imageUrl": imageURL, "updatedAt": time.Now()}, products[0])
				if err != nil {
					s.Logger.Error("Error updating product", "error", err)
					return err
				}
				continue
			}

			product := models.Product{
				SKUCode:       skuCode,
				Name:          name,
				ImageURL:      imageURL,
				PrimaryVendor: primaryVendor,
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			}

			_, err = s.ProductsRepository.Create(ctx, &product)
			if err != nil {
				s.Logger.Error("Error saving product to MongoDB", "error", err)
				return err
			}
		}

		// we can now remove the job id from cache
		err = s.TokenManager.Cache.Delete(ctx, s.ServiceCode+":"+constants.EXPORT_JOB_CODE)
		if err != nil {
			s.Logger.Error("Error deleting export job code from cache", "error", err)
			return err
		}
	}
	return nil
}

func (s *UnicommerceService) getExportJobStatus(ctx context.Context, exportJobCode string) (*ExportJobStatusResponse, error) {
	method, baseURL, path, timeout, err := s.fetchConfig(ctx, constants.API_CODE_UNICOM_EXPORT_JOB_STATUS)
	if err != nil {
		return nil, err
	}

	fullURL := baseURL + path

	payload := &ExportJobStatusPayload{
		JobCode: exportJobCode,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		s.Logger.Error("Error encoding payload for token request", "error", err)
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, strings.NewReader(string(payloadBytes)))
	if err != nil {
		s.Logger.Error("Error creating request for FetchTokens", "error", err)
		return nil, err
	}

	s.TokenManager.AuthenticateRequest(ctx, req)

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: time.Duration(timeout) * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		s.Logger.Error("Error making request to endpoint", "error", err)
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.Logger.Error("Error fetching export job status", "status", resp.StatusCode)
		return nil, err
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		s.Logger.Error("Error reading response body", "error", err)
		return nil, err
	}

	var exportJobStatusResponse ExportJobStatusResponse

	if err := json.Unmarshal(respBody, &exportJobStatusResponse); err != nil {
		s.Logger.Error("Error decoding response body", "error", err)
		return nil, err
	}

	return &exportJobStatusResponse, nil
}

// Create a function which will make a post request to unicommerce and get the inventrory snapshot, we will provide the list of SKUs
func (s *UnicommerceService) GetInventorySnapshot(ctx context.Context, skus []string) (map[string]int, error) {
	correlationID, ok := ctx.Value(constants.CorrelationID).(string)
	if !ok {
		s.Logger.Error("Correlation ID not found in context")
		return nil, errors.New("correlation ID not found in context")
	}

	method, baseURL, path, timeout, err := s.fetchConfig(ctx, constants.API_CODE_GET_INVENTORY_SNAPSHOT)
	if err != nil {
		s.Logger.Error("Error fetching config", "error", err, "correlationID", correlationID)
		return nil, err
	}

	fullURL := baseURL + path
	payload := map[string]interface{}{
		"itemTypeSKUs": skus,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		s.Logger.Error("Error encoding payload for inventory snapshot request", "error", err, "correlationID", correlationID)
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		s.Logger.Error("Error creating request for inventory snapshot", "error", err, "correlationID", correlationID)
		return nil, err
	}

	s.TokenManager.AuthenticateRequest(ctx, req)
	s.Logger.Info("Authenticated request for inventory snapshot", "correlationID", correlationID)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Facility", "Salty")
	s.Logger.Info("Set headers for request", "correlationID", correlationID)

	client := &http.Client{Timeout: time.Duration(timeout) * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		s.Logger.Error("Error making request to endpoint", "error", err, "correlationID", correlationID)
		return nil, err
	}
	s.Logger.Info("Made request to endpoint", "status", resp.StatusCode, "correlationID", correlationID)

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.Logger.Error("Error fetching inventory snapshot", "status", resp.StatusCode, "correlationID", correlationID)
		return nil, err
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		s.Logger.Error("Error reading response body", "error", err, "correlationID", correlationID)
		return nil, err
	}

	var responseData struct {
		InventorySnapshots []struct {
			ItemTypeSKU string `json:"itemTypeSKU"`
			Inventory   int    `json:"inventory"`
		} `json:"inventorySnapshots"`
	}

	err = json.Unmarshal(respBody, &responseData)
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

func (s *UnicommerceService) UpdateInventoryFromGoogleSheet(ctx context.Context) error {
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

func (s *UnicommerceService) FetchProducts(ctx context.Context) error {
	method, baseURL, path, timeout, err := s.fetchConfig(ctx, constants.API_CODE_UNICOM_FETCH_PRODUCTS)
	if err != nil {
		return err
	}

	fullURL := baseURL + path

	payload := map[string]bool{
		"getInventorySnapshot": false,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		s.Logger.Error("Error encoding payload for token request", "error", err)
		return err
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		s.Logger.Error("Error creating request for FetchTokens", "error", err)
		return err
	}

	s.TokenManager.AuthenticateRequest(ctx, req)

	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: time.Duration(timeout) * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		s.Logger.Error("Error making request to endpoint", "error", err)
		return err
	}

	defer resp.Body.Close()

	var responseData struct {
		Products []struct {
			SKUCode           string `json:"skuCode"`
			Name              string `json:"name"`
			ImageURL          string `json:"imageUrl"`
			CustomFieldValues []struct {
				FieldName  string `json:"fieldName"`
				FieldValue string `json:"fieldValue"`
			} `json:"customFieldValues"`
		} `json:"elements"`
	}

	err = json.NewDecoder(resp.Body).Decode(&responseData)
	if err != nil {
		s.Logger.Error("Error decoding response body", "error", err)
		return err
	}

	for _, p := range responseData.Products {
		product := models.Product{
			SKUCode:  p.SKUCode,
			Name:     p.Name,
			ImageURL: p.ImageURL,
		}
		for _, field := range p.CustomFieldValues {
			if field.FieldName == "Primary_Vendor" {
				product.PrimaryVendor = field.FieldValue
				break
			}
		}

		products, err := s.ProductsRepository.Find(ctx, map[string]interface{}{"skuCode": product.SKUCode, "primaryVendor": product.PrimaryVendor})

		if err != nil {
			s.Logger.Error("Error fetching products", "error", err)
			return err
		}

		if len(products) == 0 {
			product.CreatedAt = time.Now()
			product.UpdatedAt = time.Now()
			_, err = s.ProductsRepository.Create(ctx, &product)
			if err != nil {
				s.Logger.Error("Error creating product in DB", "error", err)
				return err
			}
		} else {
			// if the product already exists, we update the product
			_, err = s.ProductsRepository.Update(ctx, map[string]interface{}{"name": product.Name, "imageUrl": product.ImageURL, "updatedAt": time.Now()}, product)
			if err != nil {
				s.Logger.Error("Error updating product", "error", err)
				return err
			}
		}

	}

	return nil
}

func (s *UnicommerceService) GetProducts(ctx context.Context, skuCode string, pageNumber int, fieldsPerPage int) ([]*models.Product, int64, error) {
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
func (s *UnicommerceService) SearchProduct(ctx context.Context, skuCode string, name string) ([]*models.Product, error) {
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

// fetchFromCache retrieves the value from the cache
func (s *UnicommerceService) fetchFromCache(ctx context.Context, apiCode, key string) (string, *cache.CacheError) {
	var cacheKey string
	if apiCode == "" {
		cacheKey = s.ServiceCode + key
	} else {
		cacheKey = s.ServiceCode + ":" + apiCode + key
	}
	var value string
	err := s.TokenManager.Cache.Get(ctx, cacheKey, &value)
	if err != nil {
		return "", cache.NewCacheMissError(cacheKey)
	}
	return strings.Trim(value, "\""), nil
}

// CreatePurchaseOrder creates a new purchase order with an incremental order number
func (s *UnicommerceService) CreatePurchaseOrder(ctx context.Context, purchaseOrder *models.PurchaseOrder) error {
	// Fetch the last purchase order to determine the next order number
	lastOrder, err := s.PurchaseOrderRepository.FindOne(ctx, nil, &mongo_models.FindOptions{
		Sort: map[string]interface{}{"orderNumber": -1},
	})
	if err != nil && err != mongo_errors.ErrDocumentNotFound {
		s.Logger.Error("Error fetching last purchase order", "error", err)
		return err
	}

	// Determine the next order number
	var nextOrderNumber int
	if lastOrder != nil {
		lastOrderNumber, err := strconv.Atoi(lastOrder.PONumber)
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

	validate := validator.New()
	if err := validate.Struct(purchaseOrder); err != nil {
		s.Logger.Error("Validation error", "error", err)
		return err
	}

	// Save the purchase order to the database
	_, err = s.PurchaseOrderRepository.Create(ctx, purchaseOrder)
	if err != nil {
		s.Logger.Error("Error creating purchase order in DB", "error", err)
		return err
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
