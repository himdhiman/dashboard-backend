package services

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
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
	ServiceCode          string
	Logger               logger.ILogger
	Cache                cache.Cacher
	UnicommerceApiClient *conflux_client.ConfluxAPIClient
	ProductsRepository   *repository.Repository[models.Product]
}

func NewUnicommerceProductsService(logger logger.ILogger, cache cache.Cacher, apiClient *conflux_client.ConfluxAPIClient, productsRepository *repository.Repository[models.Product]) *UnicommerceProductsService {
	return &UnicommerceProductsService{
		ServiceCode:          products_constants.SERVICE_CODE,
		Logger:               logger,
		Cache:                cache,
		UnicommerceApiClient: apiClient,
		ProductsRepository:   productsRepository,
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

func (s *UnicommerceProductsService) AdjustUnicommerceInventory(ctx context.Context, data dto.ProductPayloadDTO) error {
	correlationID := ctx.Value(constants.CorrelationID).(string)
	s.Logger.Info("Creating purchase order", "correlationID", correlationID)

	var payload models.UnicommerceInventoryAdjustmentRequest
	payload.InventoryAdjustments = []models.UnicommerceInventoryAdjustment{
		{
			ItemSKU:   data.Data.SKU,
			Quantity:  data.Data.Quantity,
			ShelfCode: data.Data.ShelfNumber,
			AdjustmentType: func() string {
				if data.SheetName == "Sale" {
					return "REMOVE"
				}
				return "ADD"
			}(),
			Remarks:      data.Data.Remarks,
			FacilityCode: "salty",
		},
	}

	headers := map[string]string{
		"Content-Type": "application/json",
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		s.Logger.Error("Error encoding payload for token request", "error", err)
		return err
	}

	resp, err := s.UnicommerceApiClient.DoRequest(ctx, &conflux_models.APIRequest{
		ApiCode: products_constants.API_CODE_ADJUST_INVENTORY,
		Headers: headers,
		Body:    strings.NewReader(string(payloadBytes)),
	})

	if resp.StatusCode == http.StatusForbidden {
		s.Logger.Error("Received 403 Forbidden", "responseBody", string(resp.Body))
	}

	if resp.StatusCode != http.StatusOK {
		s.Logger.Error("Error creating export job", "status", resp.StatusCode)
		return err
	}

	return nil

}

func (s *UnicommerceProductsService) CreateExportJob(ctx context.Context) (*ExportJobResponse, error) {
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

	headers := map[string]string{
		"Content-Type": "application/json",
		"Facility":     "salty",
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

	err = s.Cache.Set(ctx, s.ServiceCode+":"+products_constants.EXPORT_JOB_CODE, exportJobResponse.JobCode, 0)
	if err != nil {
		s.Logger.Error("Error setting export job code in cache", "error", err)
		return nil, err
	}

	return &exportJobResponse, nil
}

// check the job status and spin the task to read the data from csv and save it in mongo
func (s *UnicommerceProductsService) CheckExportJobStatus(ctx context.Context) error {
	jobCode, err := s.FetchFromCache(ctx, products_constants.EXPORT_JOB_CODE, "")
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
		err = s.Cache.Delete(ctx, s.ServiceCode+":"+products_constants.EXPORT_JOB_CODE)
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
