package config

import (
	product_service "github.com/himdhiman/dashboard-backend/services/products-service/services"
	shipping_service "github.com/himdhiman/dashboard-backend/services/shipping-service/services"
)

type PurchaseOrderServiceConfig struct {
	ProductService  *product_service.ProductsService
	ShippingService *shipping_service.ShippingService
}

func NewPurchaseOrderServiceConfig(productsService *product_service.ProductsService,
	shippingService *shipping_service.ShippingService) *PurchaseOrderServiceConfig {
	return &PurchaseOrderServiceConfig{
		ProductService:  productsService,
		ShippingService: shippingService,
	}
}
