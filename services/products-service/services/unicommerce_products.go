package services

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/himdhiman/dashboard-backend/libs/cache"
	conflux_client "github.com/himdhiman/dashboard-backend/libs/conflux/pkg/client"
	conflux_models "github.com/himdhiman/dashboard-backend/libs/conflux/pkg/models"
	"github.com/himdhiman/dashboard-backend/libs/constants"
	"github.com/himdhiman/dashboard-backend/libs/logger"
	"github.com/himdhiman/dashboard-backend/libs/mongo/repository"
	products_constants "github.com/himdhiman/dashboard-backend/services/products-service/constants"
	"github.com/himdhiman/dashboard-backend/services/products-service/dto"
	"github.com/himdhiman/dashboard-backend/services/products-service/models"
)

type UnicommerceProductsService struct {
	ServiceCode                  string
	Logger                       logger.ILogger
	Cache                        cache.Cacher
	UnicommerceApiClient         *conflux_client.ConfluxAPIClient
	ProductsRepository           *repository.Repository[models.Product]
	ProductsBundlesRepository    *repository.Repository[models.ProductBundle]
	ShelfwiseInventoryRepository *repository.Repository[models.ShelfwiseInventory]
}

func NewUnicommerceProductsService(logger logger.ILogger, cache cache.Cacher, apiClient *conflux_client.ConfluxAPIClient, productsRepository *repository.Repository[models.Product], productsBundlesRepository *repository.Repository[models.ProductBundle], shelfwiseInventoryRepository *repository.Repository[models.ShelfwiseInventory]) *UnicommerceProductsService {
	return &UnicommerceProductsService{
		ServiceCode:                  products_constants.SERVICE_CODE,
		Logger:                       logger,
		Cache:                        cache,
		UnicommerceApiClient:         apiClient,
		ProductsRepository:           productsRepository,
		ProductsBundlesRepository:    productsBundlesRepository,
		ShelfwiseInventoryRepository: shelfwiseInventoryRepository,
	}
}

type ExportJobResponse struct {
	Successful  bool     `json:"successful"`
	Message     string   `json:"message"`
	Errors      []string `json:"errors"`
	Warnings    string   `json:"warnings"`
	ExportJobID string   `json:"exportJobId"`
	JobCode     string   `json:"jobCode"`
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

func (s *UnicommerceProductsService) AdjustUnicommerceInventory(ctx context.Context, data dto.ProductPayloadDTO) error {
	correlationID := ctx.Value(constants.CorrelationID).(string)
	s.Logger.Info("Adjusting Unicommerce Inventory", "correlationID", correlationID)

	var payload models.UnicommerceInventoryAdjustmentRequest
	payload.InventoryAdjustments = []models.UnicommerceInventoryAdjustment{
		{
			ItemSKU:   data.Data.SKU,
			Quantity:  data.Data.Quantity,
			ShelfCode: data.Data.ShelfNumber,
			InventoryType: func() string {
				if data.Data.InventoryType == "" {
					return "GOOD_INVENTORY"
				}
				return data.Data.InventoryType + "_INVENTORY"
			}(),
			AdjustmentType: func() string {
				if data.SheetName == "Sale" {
					return "REMOVE"
				}
				return "ADD"
			}(),
			Remarks:      data.Data.LotNumber + " - " + data.Data.Channel,
			FacilityCode: "Salty-HR11",
		},
	}

	s.Logger.Info("Payload for unicommerce inventory adjustment", "payload", payload, "correlationID", correlationID)

	headers := map[string]string{
		"Content-Type": "application/json",
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		s.Logger.Error("Error encoding payload for token request", "error", err, "correlationID", correlationID)
		return err
	}

	resp, err := s.UnicommerceApiClient.DoRequest(ctx, &conflux_models.APIRequest{
		ApiCode: products_constants.API_CODE_ADJUST_INVENTORY,
		Headers: headers,
		Body:    strings.NewReader(string(payloadBytes)),
	})

	if err != nil {
		s.Logger.Error("Error adjusting unicommerce inventory", "error", err, "correlationID", correlationID)
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d, response: %s", resp.StatusCode, string(resp.Body))
	}

	// Deserialize the response to model
	var response models.UnicommerceInventoryAdjustmentResponse
	err = json.Unmarshal(resp.Body, &response)
	if err != nil {
		s.Logger.Error("Error deserializing response", "error", err, "correlationID", correlationID)
		return err
	}

	s.Logger.Info("Unicommerce inventory adjustment response", "response", response, "correlationID", correlationID)

	if !response.Successful {
		s.Logger.Error("Unicommerce inventory adjustment failed", "message", response.Message, "correlationID", correlationID)
		if len(response.Errors) > 0 {
			s.Logger.Error("Unicommerce inventory adjustment errors", "errors", response.Errors, "correlationID", correlationID)
		}
		return fmt.Errorf("unicommerce inventory adjustment failed: %s", response.Message)
	}
	s.Logger.Info("Unicommerce inventory adjustment successful", "message", response.Message, "correlationID", correlationID)
	// If there are warnings, log them
	if len(response.Warnings) > 0 {
		for _, warning := range response.Warnings {
			s.Logger.Warn("Unicommerce inventory adjustment warning", "warning", warning.Message)
		}
	}

	// If there are errors in the adjustment responses, log them
	for _, adjustmentResponse := range response.InventoryAdjustmentResponses {
		if !adjustmentResponse.Successful {
			for _, err := range adjustmentResponse.Errors {
				s.Logger.Error("Unicommerce inventory adjustment response error", "error", err.Message, "correlationID", correlationID)
				return fmt.Errorf("unicommerce inventory adjustment response error: %s", err.Message)
			}
		}
	}

	s.Logger.Info("Unicommerce inventory adjustment completed successfully", "correlationID", correlationID)
	return nil

}

func (s *UnicommerceProductsService) CreateExportJobByCode(ctx context.Context, exportJobCode string) (*ExportJobResponse, error) {
	payloadFactory, ok := ExportJobPayloadFactory[exportJobCode]
	if !ok {
		s.Logger.Error("No payload factory found for export job code", "exportJobCode", exportJobCode)
		return nil, fmt.Errorf("no payload factory found for export job code: %s", exportJobCode)
	}
	payload := payloadFactory()
	return s.CreateExportJob(ctx, payload, exportJobCode)
}

func (s *UnicommerceProductsService) CreateExportJob(ctx context.Context, exportJobPayload *models.ExportJobPayload, exportJobCode string) (*ExportJobResponse, error) {
	payloadBytes, err := json.Marshal(exportJobPayload)
	if err != nil {
		s.Logger.Error("Error encoding payload for token request", "error", err)
		return nil, err
	}

	headers := map[string]string{
		"Content-Type": "application/json",
		"Facility":     "Salty-HR11",
	}

	resp, err := s.UnicommerceApiClient.DoRequest(ctx, &conflux_models.APIRequest{
		ApiCode: products_constants.API_CODE_UNICOM_CREATE_JOB,
		Headers: headers,
		Body:    strings.NewReader(string(payloadBytes)),
	})

	if resp.StatusCode != http.StatusOK {
		s.Logger.Error("Error creating export job", "status", resp.StatusCode)
		return nil, err
	}

	var exportJobResponse ExportJobResponse
	if err := json.Unmarshal(resp.Body, &exportJobResponse); err != nil {
		s.Logger.Error("Error decoding response body", "error", err)
		return nil, err
	}

	if !exportJobResponse.Successful {
		s.Logger.Error("Error creating export job", "message", exportJobResponse.Message)
		return nil, err
	}

	err = s.Cache.Set(ctx, s.ServiceCode+":"+exportJobCode, exportJobResponse.JobCode, 0)
	if err != nil {
		s.Logger.Error("Error setting export job code in cache", "error", err)
		return nil, err
	}

	return &exportJobResponse, nil
}

func (s *UnicommerceProductsService) CreateBundlesExportJob(ctx context.Context) (*ExportJobResponse, error) {
	payload := &models.ExportJobPayload{
		ExportJobTypeName: "Item Master",
		ExportColumns:     []string{"skuCode", "itemType_UDF9"},
		ExportFilters:     nil,
		Frequency:         "ONETIME",
		ReportName:        time.Now().Format("2006-01-02 15:04:05") + "_" + products_constants.BUNDLES_EXPORT_JOB_CODE,
	}

	return s.CreateExportJob(ctx, payload, products_constants.BUNDLES_EXPORT_JOB_CODE)
}

// check the job status and spin the task to read the data from csv and save it in mongo
func (s *UnicommerceProductsService) CheckExportJobStatus(ctx context.Context, exportJobCode string) error {
	jobCode, err := s.FetchFromCache(ctx, exportJobCode, "")
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

		r := csv.NewReader(bytes.NewReader(fileBytes))
		records, err := r.ReadAll()
		if err != nil {
			s.Logger.Error("Error parsing CSV file", "error", err)
			return err
		}

		processor, ok := ExportProcessorFactory[exportJobCode]
		if !ok {
			s.Logger.Error("No processor found for export job code", "exportJobCode", exportJobCode)
			return fmt.Errorf("no processor found for export job code: %s", exportJobCode)
		}

		err = processor(ctx, s, records)
		if err != nil {
			s.Logger.Error("Error processing file for export job code", "exportJobCode", exportJobCode, "error", err)
			return err
		}

		err = s.Cache.Delete(ctx, s.ServiceCode+":"+exportJobCode)
		if err != nil {
			s.Logger.Error("Error deleting export job code from cache", "error", err)
			return err
		}
	}
	return nil
}

func (s *UnicommerceProductsService) getExportJobStatus(ctx context.Context, exportJobCode string) (*ExportJobStatusResponse, error) {
	payload := &ExportJobStatusPayload{
		JobCode: exportJobCode,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		s.Logger.Error("Error encoding payload for token request", "error", err)
		return nil, err
	}

	headers := map[string]string{
		"Content-Type": "application/json",
	}

	resp, err := s.UnicommerceApiClient.DoRequest(ctx, &conflux_models.APIRequest{
		ApiCode: products_constants.API_CODE_UNICOM_EXPORT_JOB_STATUS,
		Headers: headers,
		Body:    strings.NewReader(string(payloadBytes)),
	})

	if resp.StatusCode != http.StatusOK {
		s.Logger.Error("Error fetching export job status", "status", resp.StatusCode)
		return nil, err
	}

	var exportJobStatusResponse ExportJobStatusResponse

	if err := json.Unmarshal(resp.Body, &exportJobStatusResponse); err != nil {
		s.Logger.Error("Error decoding response body", "error", err)
		return nil, err
	}

	return &exportJobStatusResponse, nil
}

func (s *UnicommerceProductsService) FetchFromCache(ctx context.Context, apiCode, key string) (string, *cache.CacheError) {
	var cacheKey string
	if apiCode == "" {
		cacheKey = s.ServiceCode + key
	} else {
		cacheKey = s.ServiceCode + ":" + apiCode + key
	}
	var value string
	err := s.Cache.Get(ctx, cacheKey, &value)
	if err != nil {
		return "", cache.NewCacheMissError(cacheKey)
	}
	return strings.Trim(value, "\""), nil
}

// CheckJobStatusByJobCode checks if a job is running or finished based on the jobCode
// Returns "running" if job is still being processed internally, "finished" if processing is complete
// This checks our internal processing status, not the Unicommerce API status
func (s *UnicommerceProductsService) CheckJobStatusByJobCode(ctx context.Context, jobCode string) (string, error) {
	correlationID := ctx.Value(constants.CorrelationID).(string)
	s.Logger.Info("Checking internal job processing status by job code", "jobCode", jobCode, "correlationID", correlationID)

	// Check all possible export job codes to find if this jobCode is still in cache
	// The cache stores: ServiceCode:exportJobCode -> jobCode
	// If the key exists, processing is still ongoing. If deleted, processing is finished.
	exportJobCodes := []string{
		products_constants.PRODUCTS_EXPORT_JOB_CODE,
		products_constants.BUNDLES_EXPORT_JOB_CODE,
		products_constants.SHELFWISE_INVENTORY_EXPORT_JOB_CODE,
	}

	for _, exportJobCode := range exportJobCodes {
		cachedJobCode, err := s.FetchFromCache(ctx, exportJobCode, "")
		if err == nil && cachedJobCode == jobCode {
			// Found the jobCode in cache, meaning processing is still ongoing
			s.Logger.Info("Job found in cache, processing still running", "jobCode", jobCode, "exportJobCode", exportJobCode, "correlationID", correlationID)
			return "running", nil
		}
	}

	// JobCode not found in any cache key, meaning processing is finished
	s.Logger.Info("Job not found in cache, processing finished", "jobCode", jobCode, "correlationID", correlationID)
	return "finished", nil
}
