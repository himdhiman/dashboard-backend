package config

import (
	product_service "github.com/himdhiman/dashboard-backend/services/products-service/services"
)

type ShippingServiceConfig struct {
	ProductService *product_service.ProductsService
}

func NewShippingServiceConfig(productsService *product_service.ProductsService) *ShippingServiceConfig {
	return &ShippingServiceConfig{
		ProductService: productsService,
	}
}
