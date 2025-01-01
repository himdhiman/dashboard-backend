package services

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/himdhiman/dashboard-backend/libs/logger"
	mongo_models "github.com/himdhiman/dashboard-backend/libs/mongo/models"
	"github.com/himdhiman/dashboard-backend/libs/mongo/repository"
	"github.com/himdhiman/dashboard-backend/services/sentinel-service/auth"
	"github.com/himdhiman/dashboard-backend/services/sentinel-service/constants"
	"github.com/himdhiman/dashboard-backend/services/sentinel-service/models"
)

type UnicommerceService struct {
	ServiceCode        string
	Logger             logger.LoggerInterface
	TokenManager       *auth.TokenManager
	ProductsRepository *repository.Repository[models.Product]
}

func NewUnicommerceService(tokenManager *auth.TokenManager, logger logger.LoggerInterface, productsCollection *mongo_models.MongoCollection) *UnicommerceService {

	productsRepo := repository.Repository[models.Product]{Collection: productsCollection}

	return &UnicommerceService{
		ServiceCode:        constants.UNICOM_API_CODE,
		TokenManager:       tokenManager,
		Logger:             logger,
		ProductsRepository: &productsRepo,
	}
}

func (s *UnicommerceService) GetItemType(ctx context.Context, skuCode string) (*http.Response, error) {
	payload := map[string]string{
		"skuCode": skuCode,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		s.Logger.Error("Error encoding payload for token request", "error", err)
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://salty.unicommerce.com/services/rest/v1/catalog/itemType/get", bytes.NewBuffer(payloadBytes))
	if err != nil {
		s.Logger.Error("Error creating request for FetchTokens", "error", err)
		return nil, err
	}

	s.TokenManager.AuthenticateRequest(ctx, req)

	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		s.Logger.Error("Error making request to endpoint", "error", err)
		return nil, err
	}

	return resp, nil
}

func (s *UnicommerceService) FetchProducts(ctx context.Context) error {
	method, err := s.fetchFromCache(ctx, constants.API_CODE_UNICOM_FETCH_PRODUCTS, constants.API_METHOD)
	if err != nil {
		s.Logger.Error("Error fetching method from cache", "error", err)
		return err
	}

	baseURL, err := s.fetchFromCache(ctx, "", constants.BASE_URL)
	if err != nil {
		s.Logger.Error("Error fetching base URL from cache", "error", err)
		return err
	}

	path, err := s.fetchFromCache(ctx, constants.API_CODE_UNICOM_FETCH_PRODUCTS, constants.API_PATH)
	if err != nil {
		s.Logger.Error("Error fetching path from cache", "error", err)
		return err
	}

	timeoutStr, err := s.fetchFromCache(ctx, constants.API_CODE_UNICOM_FETCH_PRODUCTS, constants.API_TIMEOUT)
	if err != nil {
		s.Logger.Error("Error fetching timeout from cache", "error", err)
		return err
	}
	timeout, err := strconv.Atoi(timeoutStr)
	if err != nil {
		s.Logger.Error("Error converting timeout to integer", "error", err)
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
			ID                int    `json:"id"`
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
			ID:       p.ID,
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
			_, err = s.ProductsRepository.Create(ctx, &product)
			if err != nil {
				s.Logger.Error("Error creating product in DB", "error", err)
				return err
			}
		} else {
			// if the product already exists, we update the product
			_, err = s.ProductsRepository.Update(ctx, map[string]interface{}{"name": product.Name, "imageUrl": product.ImageURL}, product)
			if err != nil {
				s.Logger.Error("Error updating product", "error", err)
				return err
			}
		}

	}

	return nil
}

// we will fetch all the products from the database
// if skuCode is provided, we will filter the products by skuCode
// we will also take the page number and number of fields per page and fetch the products accordingly
func (s *UnicommerceService) GetProducts(ctx context.Context, skuCode string, pageNumber int, fieldsPerPage int) ([]*models.Product, error) {
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
		return nil, err
	}

	return products, nil
}

// fetchFromCache retrieves the value from the cache
func (s *UnicommerceService) fetchFromCache(ctx context.Context, apiCode, key string) (string, error) {
	var cacheKey string
	if apiCode == "" {
		cacheKey = s.ServiceCode + key
	} else {
		cacheKey = s.ServiceCode + ":" + apiCode + key
	}
	var value string
	err := s.TokenManager.Cache.Get(ctx, cacheKey, &value)
	if err != nil {
		return "", err
	}
	return strings.Trim(value, "\""), nil
}
