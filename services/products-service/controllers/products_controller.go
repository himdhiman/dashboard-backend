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
		c.JSON(http.StatusBadRequest, gin.H{"error": constants.ErrMissingCorrelationID})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	products, err := uc.Service.SearchProduct(ctx, request.SKUCode, request.Name)
	if err != nil {
		uc.Logger.Error("Error searching products", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search products"})
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

	c.JSON(http.StatusOK, gin.H{"products": response})
}
