package controllers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/himdhiman/dashboard-backend/libs/constants"
	"github.com/himdhiman/dashboard-backend/libs/logger"
	"github.com/himdhiman/dashboard-backend/services/products-service/models"
	"github.com/himdhiman/dashboard-backend/services/products-service/services"
)

type ProductsController struct {
	Logger  logger.ILogger
	Service *services.ProductsService
}

func NewProductsController(logger logger.ILogger, service *services.ProductsService) *ProductsController {
	return &ProductsController{
		Logger:  logger,
		Service: service,
	}
}

type GetProductsResponse struct {
	Data  []models.Product `json:"data"`
	Total int              `json:"total"`
	Page  int              `json:"page"`
	Limit int              `json:"limit"`
}

func (uc *ProductsController) GetProducts(c *gin.Context) {
	ctx := c.Request.Context()
	correlationID := c.GetHeader(string(constants.CorrelationID))
	if correlationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": constants.ErrMissingCorrelationID})
		return
	}
	ctx = context.WithValue(ctx, constants.CorrelationID, correlationID)
	// Parse query parameters

	uc.Logger.Info("Getting products", "correlationID", correlationID)

	pageNumberStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")
	skuCode := c.DefaultQuery("skuCode", "")

	pageNumber, err := strconv.Atoi(pageNumberStr)
	if err != nil || pageNumber < 1 {
		pageNumber = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 10
	}

	/// Fetch products
	productsPtr, total, err := uc.Service.GetProducts(ctx, skuCode, pageNumber, limit)
	if err != nil {
		uc.Logger.Error("Error fetching products", "correlationID", correlationID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch products"})
		return
	}

	products := make([]models.Product, len(productsPtr))
	for i, p := range productsPtr {
		products[i] = *p
	}

	response := GetProductsResponse{
		Data:  products,
		Total: int(total),
		Page:  pageNumber,
		Limit: limit,
	}

	c.JSON(http.StatusOK, response)
}

func (uc *ProductsController) SearchProduct(c *gin.Context) {
	ctx := c.Request.Context()
	correlationID := c.GetHeader(string(constants.CorrelationID))
	if correlationID == "" {
		uc.Logger.Error("Missing correlation ID")
		respondWithError(c, http.StatusBadRequest, "Missing correlation ID", constants.ErrMissingCorrelationID)
		return
	}
	ctx = context.WithValue(ctx, constants.CorrelationID, correlationID)

	var request struct {
		SKUCode string   `json:"skuCode"`
		Name    string   `json:"name"`
		Fields  []string `json:"fields"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		uc.Logger.Error("Error binding JSON", "error", err)
		respondWithError(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	products, err := uc.Service.SearchProduct(ctx, request.SKUCode, request.Name)
	if err != nil {
		uc.Logger.Error("Error searching products", "error", err)
		respondWithError(c, http.StatusInternalServerError, "Failed to search products", err.Error())
		return
	}

	response := make([]map[string]interface{}, len(products))
	for i, product := range products {
		productMap := make(map[string]interface{})
		for _, field := range request.Fields {
			switch field {
			case "id":
				productMap["id"] = product.ID
			case "name":
				productMap["name"] = product.Name
			case "sku":
				productMap["sku"] = product.SKUCode
			case "imageUrl":
				productMap["imageUrl"] = product.ImageURL
			case "primaryVendor":
				productMap["primaryVendor"] = product.PrimaryVendor
			case "lastProcuredRmbPrice":
				productMap["lastProcuredRmbPrice"] = product.LastProcuredRmbPrice
			case "createdAt":
				productMap["createdAt"] = product.CreatedAt
			case "updatedAt":
				productMap["updatedAt"] = product.UpdatedAt
			default:
				uc.Logger.Warn("Unknown field requested", "field", field)
			}
		}
		response[i] = productMap
	}

	respondWithSuccess(c, http.StatusOK, "Products fetched successfully", response)
}

// GetProductBundles retrieves product bundles with pagination
func (uc *ProductsController) GetProductBundles(c *gin.Context) {
	uc.Logger.Info("Getting product bundles")
	ctx := c.Request.Context()
	correlationID := c.GetHeader(string(constants.CorrelationID))
	if correlationID == "" {
		uc.Logger.Error("Missing correlation ID")
		respondWithError(c, http.StatusBadRequest, "Missing correlation ID", constants.ErrMissingCorrelationID)
		return
	}
	ctx = context.WithValue(ctx, constants.CorrelationID, correlationID)

	// Parse query parameters
	pageNumberStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")
	bundleSKU := c.DefaultQuery("bundleSku", "")

	pageNumber, err := strconv.Atoi(pageNumberStr)
	if err != nil || pageNumber < 1 {
		pageNumber = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 10
	}

	// Fetch product bundles
	bundlesPtr, total, err := uc.Service.GetProductBundles(ctx, bundleSKU, pageNumber, limit)
	if err != nil {
		uc.Logger.Error("Error fetching product bundles", "correlationID", correlationID, "error", err)
		respondWithError(c, http.StatusInternalServerError, "Failed to fetch product bundles", err)
		return
	}

	bundles := make([]models.ProductBundle, len(bundlesPtr))
	for i, b := range bundlesPtr {
		bundles[i] = *b
	}

	response := struct {
		Data  []models.ProductBundle `json:"data"`
		Total int                    `json:"total"`
		Page  int                    `json:"page"`
		Limit int                    `json:"limit"`
	}{
		Data:  bundles,
		Total: int(total),
		Page:  pageNumber,
		Limit: limit,
	}

	uc.Logger.Info("Successfully fetched product bundles", "correlationID", correlationID)
	respondWithSuccess(c, http.StatusOK, "Product bundles fetched successfully", response)
}

func (uc *ProductsController) GetProductBundleByID(c *gin.Context) {
	ctx := c.Request.Context()
	correlationID := c.GetHeader(string(constants.CorrelationID))
	if correlationID == "" {
		uc.Logger.Error("Missing correlation ID")
		respondWithError(c, http.StatusBadRequest, "Missing correlation ID", constants.ErrMissingCorrelationID)
		return
	}
	ctx = context.WithValue(ctx, constants.CorrelationID, correlationID)
	
	bundleID := c.Param("id")
	if bundleID == "" {
		uc.Logger.Error("Missing bundle ID")
		respondWithError(c, http.StatusBadRequest, "Missing bundle ID", "bundle ID is required")
		return
	}

	bundle, err := uc.Service.GetProductBundleByID(ctx, bundleID)
	if err != nil {
		uc.Logger.Error("Error fetching product bundle", "error", err)
		respondWithError(c, http.StatusInternalServerError, "Failed to fetch product bundle", err.Error())
		return
	}
	if bundle == nil {
		uc.Logger.Warn("Product bundle not found", "bundleID", bundleID)
		respondWithError(c, http.StatusNotFound, "Product bundle not found", "No bundle found with the given ID")
		return
	}

	respondWithSuccess(c, http.StatusOK, "Product bundle fetched successfully", bundle)
}

// func (uc *ProductsController) GetShelfwiseInventory(c *gin.Context) {
// 	ctx := c.Request.Context()
// 	correlationID := c.GetHeader(string(constants.CorrelationID))
// 	if correlationID == "" {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": constants.ErrMissingCorrelationID})
// 		return
// 	}
// 	ctx = context.WithValue(ctx, constants.CorrelationID, correlationID)

// 	// Parse query parameters
// 	pageNumberStr := c.DefaultQuery("page", "1")
// 	limitStr := c.DefaultQuery("limit", "10")
// 	skuCode := c.DefaultQuery("skuCode", "")
// 	facilityCode := c.DefaultQuery("facilityCode", "")

// 	pageNumber, err := strconv.Atoi(pageNumberStr)
// 	if err != nil || pageNumber < 1 {
// 		pageNumber = 1
// 	}

// 	limit, err := strconv.Atoi(limitStr)
// 	if err != nil || limit < 1 {
// 		limit = 10
// 	}

// 	// Fetch shelfwise inventory
// 	inventoryPtr, total, err := uc.Service.GetShelfwiseInventory(ctx, skuCode, facilityCode, pageNumber, limit)
// 	if err != nil {
// 		uc.Logger.Error("Error fetching shelfwise inventory", "correlationID", correlationID, "error", err)
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch shelfwise inventory"})
// 		return
// 	}

// 	inventory := make([]models.ShelfwiseInventory, len(inventoryPtr))
// 	for i, inv := range inventoryPtr {
// 		inventory[i] = *inv
// 	}

// 	response := struct {
// 		Data  []models.ShelfwiseInventory `json:"data"`
// 		Total int                         `json:"total"`
// 		Page  int                         `json:"page"`
// 		Limit int                         `json:"limit"`
// 	}{
// 		Data:  inventory,
// 		Total: int(total),
// 		Page:  pageNumber,
// 		Limit: limit,
// 	}

// 	c.JSON(http.StatusOK, response)
// }
